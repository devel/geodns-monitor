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
	Ip               string    `json:"ip"`
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
	nextServerId  chan int
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
	hub.nextServerId = make(chan int)
	hub.configManager = make(chan bool)
	go hub.makeServerId()
	go hub.arbiter()
	return hub
}

func (s *StatusHub) MarkConfigurationStart() {
	s.configManager <- false
}

func (s *StatusHub) MarkConfigurationEnd() {
	s.configManager <- true
}

func (s *StatusHub) makeServerId() int {
	i := 1
	for {
		log.Println("Ready to make server id", i)
		s.nextServerId <- i
		i++
	}
}

func (s *StatusHub) arbiter() {
	log.Println("running arbiter")
	for {
		select {
		case new := <-s.statusUpdates:
			// log.Println("Adding status for", new.Ip)
			srv := s.serverStatus[new.connId]
			updateStatus(srv, new)
			// TODO: push to seriesly

		case msg := <-s.statusMsgChan:
			// log.Printf("Got StatusMsg from '%s': %s\n", msg.connId, msg.Status)
			s.serverStatus[msg.connId].Status = msg.Status

		case s.statuses <- s.serverStatus:

		case cm := <-s.configManager:
			switch cm {
			case false:
				s.configRevision++
			case true:
				for connId, srv := range s.serverStatus {
					if srv.Connection.configRevision < s.configRevision {
						log.Printf("Server %s has an old config revision, disconnecting %d", srv.Ip, connId)
						srv.Connection.Stop()
						// delete(s.serverStatus, connId)
					}
				}
			}

		case ip := <-s.addServerChan:

			log.Println("Adding monitoring of", ip)

			for _, server := range s.serverStatus {
				if server.Ip == ip.String() {
					log.Printf("Already monitoring '%s'\n", ip.String())
					continue
				}
			}

			log.Printf("Creating new connection for %s", ip)

			sc := NewServerConnection(ip, s.statusUpdates, s.statusMsgChan)
			sc.configRevision = s.configRevision

			log.Printf("Start() on %s", sc.Ip)

			connId := <-s.nextServerId

			log.Println("got server id", connId)

			status := new(Status)
			status.Ip = ip.String()
			status.Connection = sc
			s.serverStatus[connId] = status

			sc.Start(connId)

		case <-s.quit:
			log.Printf("StatusHub got quit!\n")
			for connId, srv := range s.serverStatus {
				log.Printf("Sending quit to %d (%s)\n", connId, srv.Ip)
				srv.Connection.Stop()
				delete(s.serverStatus, connId)
			}
			// TODO: do we need to close the channels?
			log.Println("Arbiter done")
			return
		}
	}
}

func updateStatus(srv *Status, new *ServerUpdate) {
	srv.Data = *new
	srv.LastStatusUpdate = time.Now()

	if len(new.Version) > 0 {
		srv.Version = new.Version
	}

	if len(new.Id) > 0 {
		srv.Name = new.Id
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
	rv := make([]*Status, len(current))
	i := 0
	for _, status := range current {
		// log.Printf("Status for '%name': %#v\n", name, status)
		rv[i] = status
		i++
	}
	return rv
}

func (s *StatusHub) Stop() {
	log.Println("Sending quit to hub")
	s.quit <- true
	log.Println("sent quit to hub")
}

func (s *StatusHub) addIp(ip net.IP) error {
	s.addServerChan <- ip
	return nil
}

func (s *StatusHub) AddName(ipstr string) error {
	ip := net.ParseIP(ipstr)
	if ip == nil {
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
			err = s.addIp(addr)
			if err != nil {
				log.Printf("Could not add '%s': %s\n", addr, err)
			}
		}
		return nil
	} else {
		return s.addIp(ip)
	}
}
