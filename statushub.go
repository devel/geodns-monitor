package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Status struct {
	Name             string    `json:"name"`
	Names            []string  `json:"names"`
	Groups           []string  `json:"groups"`
	IP               string    `json:"ip"`
	UUID             string    `json:"uuid"`
	Version          string    `json:"version"`
	Queries          int64     `json:"queries"`
	Qps              float64   `json:"qps"`
	Qps1             float64   `json:"qps1m"`
	Uptime           int64     `json:"uptime"`
	Status           string    `json:"status"`
	LastStatusUpdate time.Time `json:"-"`

	Connection *ServerConnection

	Data ServerUpdate
}

type statusMap map[int]*Status

type StatusHub struct {
	statusUpdates chan *ServerUpdate
	statusMsgChan chan *ServerStatusMsg
	addServerChan chan net.IP
	nextServerID  chan int
	serverStatus  statusMap
	statuses      chan statusMap
	remove        chan string
	quit          chan bool

	configRevision int
	configManager  chan bool
}

func NewHub() *StatusHub {
	hub := new(StatusHub)
	hub.statusUpdates = make(chan *ServerUpdate, 10)
	hub.statusMsgChan = make(chan *ServerStatusMsg, 10)
	hub.addServerChan = make(chan net.IP)
	hub.statuses = make(chan statusMap)
	hub.quit = make(chan bool, 1)
	hub.serverStatus = make(statusMap)
	hub.nextServerID = make(chan int)
	hub.configManager = make(chan bool)
	go hub.makeServerID()
	go hub.arbiter()
	return hub
}

func (s *StatusHub) MarkConfigurationStart() {
	s.configManager <- false
}

func (s *StatusHub) MarkConfigurationEnd() {
	s.configManager <- true
}

func (s *StatusHub) makeServerID() int {
	i := 1
	for {
		log.Println("Ready to make server id", i)
		s.nextServerID <- i
		i++
	}
}

func (s *StatusHub) arbiter() {
	log.Println("running arbiter")
	for {
		select {
		case new := <-s.statusUpdates:
			// log.Println("Adding status for", new.IP)
			srv, ok := s.serverStatus[new.ConnID]
			if ok {
				if len(new.UUID) > 0 {
					if dupeID := s.FindUUID(new.UUID); dupeID > 0 && dupeID != new.ConnID {
						log.Printf("Duplicate connection to %s (uuid %s); this is %d, dupe is %d", new.IP, new.UUID, new.ConnID, dupeID)

						// try keeping the connection that's the one reported by the server
						if srv.Connection.IP.String() != new.IP {
							dupeID = new.ConnID
						}
						s.serverStatus[dupeID].Connection.Stop()
						delete(s.serverStatus, dupeID)
						continue
					}
				}

				updateStatus(srv, new)
			} else {
				log.Printf("got status update for unknown connection %d (ip %s)", new.ConnID, new.IP)
			}

			// TODO: push to seriesly

		case msg := <-s.statusMsgChan:
			// log.Printf("Got StatusMsg from '%d': %s\n", msg.ConnID, msg.Status)
			srv, ok := s.serverStatus[msg.ConnID]
			if ok {
				srv.Status = msg.Status
			}

		case s.statuses <- s.serverStatus:

		case cm := <-s.configManager:
			switch cm {
			case false:
				s.configRevision++
			case true:
				for connID, srv := range s.serverStatus {
					if srv.Connection.configRevision < s.configRevision {
						log.Printf("Server %s has an old config revision, disconnecting %d", srv.IP, connID)
						srv.Connection.Stop()
						delete(s.serverStatus, connID)
					}
				}
			}

		case ip := <-s.addServerChan:

			log.Println("Adding monitoring of", ip)

			foundDuplicate := false
			for _, server := range s.serverStatus {
				if server.IP == ip.String() {
					foundDuplicate = true
					log.Printf("Already monitoring '%s'\n", ip.String())
					server.Connection.configRevision = s.configRevision
					break
				}
			}
			if foundDuplicate {
				continue
			}

			log.Printf("Creating new connection for %s", ip)

			sc := NewServerConnection(ip, s.statusUpdates, s.statusMsgChan)
			sc.configRevision = s.configRevision

			log.Printf("Start() on %s", sc.IP)

			connID := <-s.nextServerID

			log.Println("got server id", connID)

			status := new(Status)
			status.IP = ip.String()
			status.Connection = sc
			s.serverStatus[connID] = status

			sc.Start(connID)

		case <-s.quit:
			log.Printf("StatusHub got quit!\n")
			for connID, srv := range s.serverStatus {
				log.Printf("Sending quit to %d (%s)\n", connID, srv.IP)
				srv.Connection.Stop()
				delete(s.serverStatus, connID)
			}
			// TODO: do we need to close the channels?
			log.Println("Arbiter done")
			return
		}
	}
}

func (s *StatusHub) FindUUID(UUID string) int {
	for connID, server := range s.serverStatus {
		if server.UUID == UUID {
			return connID
		}
	}
	return 0
}

func updateStatus(srv *Status, new *ServerUpdate) {
	srv.Data = *new
	srv.LastStatusUpdate = time.Now()

	if len(new.Version) > 0 {
		srv.Version = new.Version
	}

	if len(new.ID) > 0 {
		srv.Name = new.ID
	}

	if len(new.UUID) > 0 {
		srv.UUID = new.UUID
	}

	if new.Uptime > 0 {
		srv.Uptime = new.Uptime
	}

	srv.Qps = new.Qps
	srv.Queries = new.Queries

	if new.Qps1 > 0 {
		srv.Qps1 = new.Qps1
	}

	if len(new.Hostname) > 0 {
		// This needs to accumulate the various names that have been
		// discovered for this server, maybe.
		srv.Names = []string{new.Hostname}
	}

	if len(new.Groups) > 0 {
		srv.Groups = new.Groups
	}

}

func (s *StatusHub) Status() []*Status {
	current := <-s.statuses
	rv := make([]*Status, 0)
	for _, status := range current {
		// log.Printf("Status for '%name': %#v\n", name, status)
		if !status.LastStatusUpdate.IsZero() || len(status.Status) > 0 {
			rv = append(rv, status)
		}
	}
	return rv
}

func (s *StatusHub) Stop() {
	s.quit <- true
}

func (s *StatusHub) addIP(ip net.IP) error {
	s.addServerChan <- ip
	return nil
}

func (s *StatusHub) AddNameBackground(ipstr string, ch chan error) {
	go func() {
		err := s.AddName(ipstr)
		if err == nil {
			ch <- err
		} else {
			ch <- fmt.Errorf("error adding server '%s': %s", ipstr, err)
		}
	}()
}

func (s *StatusHub) AddName(ipstr string) error {
	ip := net.ParseIP(ipstr)
	if ip != nil {
		return s.addIP(ip)
	}
	// return fmt.Errorf("Could not parse IP: '%s'", ipstr)
	addrs, err := net.LookupIP(ipstr)
	log.Printf("IP: %s, %#v %d\n", ipstr, addrs, len(addrs))
	if err != nil || len(addrs) == 0 {
		return fmt.Errorf("Could not lookup name: '%s': %s", ipstr, err)
	}

	if false {
		return fmt.Errorf("Could not find IPs for: '%s'\n", ipstr)
	}

	for _, addr := range addrs {
		log.Println("Adding", addr)
		err = s.addIP(addr)
		if err != nil {
			log.Printf("Could not add '%s': %s\n", addr, err)
		}
	}
	return nil
}
