package main

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

type ServerStatusMsg struct {
	connId int
	Status string
}

type ServerConnection struct {
	connId        int
	Ip            net.IP
	updateChan    chan *ServerUpdate
	statusMsgChan chan *ServerStatusMsg
	quit          chan bool
}

type ServerUpdate struct {
	connId   int
	Hostname string   `json:"h"`
	Id       string   `json:"id"`
	Version  string   `json:"v"`
	Groups   []string `json:"groups"`
	Ip       string   `json:"ip"`
	Uptime   int64    `json:"up"`
	Queries  int64    `json:"qs"`
	Qps      float64  `json:"qps"`
	Qps1     float64  `json:"qps1m"`
	Started  int      `json:"started"`
}

func NewServerConnection(ip net.IP, updates chan *ServerUpdate, sm chan *ServerStatusMsg) *ServerConnection {
	sc := new(ServerConnection)
	sc.Ip = ip
	sc.updateChan = updates
	sc.statusMsgChan = sm
	sc.quit = make(chan bool)
	return sc
}

func (sc *ServerConnection) Start(id int) {
	sc.connId = id
	q := sc.quit
	c := sc.updateChan
	sm := sc.statusMsgChan
	sc.statusMsg("Starting")

	su := new(ServerUpdate)
	su.connId = id
	su.Ip = sc.Ip.String()

	c <- su

	go sc.start(c, sm, q)
}

func (sc *ServerConnection) statusErrorMsg(str string) {
	sc.statusMsg(str)
	su := new(ServerUpdate)
	su.connId = sc.connId
	su.Ip = sc.Ip.String()
	sc.updateChan <- su
}

func (sc *ServerConnection) statusMsg(str string) {
	msg := &ServerStatusMsg{sc.connId, str}
	sc.statusMsgChan <- msg
}

func (sc *ServerConnection) start(c chan *ServerUpdate, msg chan *ServerStatusMsg, q chan bool) {
	log.Println("Fetch for", sc.Ip)

	retries := 0

	for {

		conn, err := net.Dial("tcp", net.JoinHostPort(sc.Ip.String(), "8053"))
		if err != nil {
			status := fmt.Sprintf("Could not connect to '%s': %s", sc.Ip, err)
			sc.statusErrorMsg(status)
			log.Println(status)
			retries++
			time.Sleep(time.Second * 8)
			continue
		}
		url, err := url.Parse("/monitor")
		if err != nil {
			log.Println("Could not parse url", err)
		}
		header := http.Header{}
		header.Add("Origin", "http://monitor.pgeodns")
		header.Add("Host", sc.Ip.String())
		header.Add("Set-WebSocket-Protocol", "chat")

		ws, resp, err := websocket.NewClient(conn, url, header, 1024, 1024)
		if err != nil {
			status := fmt.Sprintf("Could not upgrade WS on '%s': %s", sc.Ip, err)
			sc.statusErrorMsg(status)
			log.Println(status)
			retries++
			time.Sleep(time.Second * 1)
			continue
		}
		retries = 0
		log.Println("Response", resp)

		status := new(ServerUpdate)
		status.connId = sc.connId

		for {
			ws.SetReadDeadline(time.Now().Add(time.Second * 3))
			sc.statusMsg("Ok")
			op, r, err := ws.NextReader()
			if err != nil {
				status := fmt.Sprintf("Error reading from server: %s", err)
				sc.statusErrorMsg(status)
				log.Println(status)
				break
			}
			msg, err := ioutil.ReadAll(r)

			// log.Println("op", op, "msg", string(msg), "err", err)

			if op == websocket.OpText {
				err = json.Unmarshal(msg, &status)
				if err != nil {
					log.Printf("Unmarshall err from '%s': '%s', data: '%s'\n", sc.Ip.String(), err, msg)
				}
				// log.Printf("Got status: %#v\n", status)
				c <- status
			} else {
				log.Println("op", op, "msg", string(msg), "err", err)
			}

			// os.Exit(0)
		}
	}
}
