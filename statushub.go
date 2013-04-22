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

	Data ServerUpdate
}

type statusMap map[int]*Status

type StatusHub struct {
	statusUpdates chan *ServerUpdate
	statusMsgChan chan *ServerStatusMsg
	nextServerId  chan int
	Servers       []ServerConnection
	serverStatus  statusMap
	statuses      chan statusMap
	remove        chan string
	quit          chan bool
}

func NewHub() *StatusHub {
	hub := new(StatusHub)
	hub.statusUpdates = make(chan *ServerUpdate, 10)
	hub.statusMsgChan = make(chan *ServerStatusMsg, 10)
	hub.statuses = make(chan statusMap)
	hub.quit = make(chan bool)
	hub.serverStatus = make(statusMap)
	hub.nextServerId = make(chan int)
	go hub.makeServerId()
	go hub.arbiter()
	return hub
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

		case <-s.quit:
			log.Printf("StatusHub got quit!\n")
			for _, sc := range s.Servers {
				log.Printf("Sending quit to %s\n", sc.Ip)
				delete(s.serverStatus, sc.connId)
				sc.quit <- true
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
	s.quit <- true
}

func (s *StatusHub) addIp(ip net.IP) error {

	log.Printf("Creating new connection for %s", ip)

	sc := NewServerConnection(ip, s.statusUpdates, s.statusMsgChan)

	log.Printf("Start() on %s", sc.Ip)

	connId := <-s.nextServerId

	log.Println("got server id", connId)

	status := new(Status)
	status.Ip = ip.String()
	s.serverStatus[connId] = status

	sc.Start(connId)

	log.Println("Add() returning")

	return nil
}

func (s *StatusHub) AddName(ipstr string) error {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		// return fmt.Errorf("Could not parse IP: '%s'", ipstr)
		addrs, err := net.LookupIP(ipstr)
		if err != nil {
			return fmt.Errorf("Could not lookup name: '%s': %s", ipstr, err)
		}
		for _, addr := range addrs {
			log.Println("Adding", addr)
			s.addIp(addr)
		}
		return nil
	} else {
		return s.addIp(ip)
	}
}
