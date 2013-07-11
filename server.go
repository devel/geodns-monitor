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

	quit           chan bool
	sleep          chan int
	configRevision int
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
	sc.quit = make(chan bool, 1)
	sc.sleep = make(chan int, 1)
	return sc
}

func (sc *ServerConnection) Start(id int) {
	sc.connId = id
	sc.statusMsg("Starting")

	su := new(ServerUpdate)
	su.connId = id
	su.Ip = sc.Ip.String()

	sc.updateChan <- su

	go sc.start()
}

func (sc *ServerConnection) Stop() {
	sc.quit <- true
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

func (sc *ServerConnection) start() {
	log.Println("Fetch for", sc.Ip)

	retries := 0

	for {

		select {

		case <-sc.quit:
			log.Println("sc got quit!")
			sc.statusErrorMsg("stopped")
			return

		case retries := <-sc.sleep:
			delay := retries * retries / 2
			if delay > 60 {
				delay = 30
			}
			time.Sleep(time.Duration(delay) * time.Second)

		default:

			retries++

			conn, err := net.Dial("tcp", net.JoinHostPort(sc.Ip.String(), "8053"))
			if err != nil {
				status := fmt.Sprintf("%s", err)
				sc.statusErrorMsg(status)
				log.Println(status)
				sc.sleep <- retries
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

			ws, _, err := websocket.NewClient(conn, url, header, 1024, 1024)
			if err != nil {
				status := fmt.Sprintf("Could not upgrade WS on '%s': %s", sc.Ip, err)
				sc.statusErrorMsg(status)
				log.Println(status)
				sc.sleep <- retries
				continue
			}
			sc.read(ws)
			log.Println("server reader stopped")
			err = conn.Close()
			if err != nil {
				log.Printf("Error closing connection to %s: %s", sc.Ip, err)
			}
			sc.sleep <- retries
			continue
		}
	}
}

func (sc *ServerConnection) read(ws *websocket.Conn) {

	// log.Println("Response", resp)

	status := new(ServerUpdate)
	status.connId = sc.connId

	for {

		select {
		case <-sc.quit:
			log.Println("server reader got quit message")
			sc.quit <- true
			return

		default:

			ws.SetReadDeadline(time.Now().Add(time.Second * 3))
			sc.statusMsg("Ok")
			op, r, err := ws.NextReader()
			if err != nil {
				status := fmt.Sprintf("Error reading from server: %s", err)
				sc.statusErrorMsg(status)
				log.Println(status)
				return
			}
			msg, err := ioutil.ReadAll(r)

			// log.Println("op", op, "msg", string(msg), "err", err)

			if op == websocket.OpText {
				err = json.Unmarshal(msg, &status)
				if err != nil {
					log.Printf("Unmarshall err from '%s': '%s', data: '%s'\n", sc.Ip.String(), err, msg)
				}
				// log.Printf("Got status: %#v\n", status)
				sc.updateChan <- status
			} else {
				log.Println("op", op, "msg", string(msg), "err", err)
			}

			// os.Exit(0)
		}
	}
}
