package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
)

var VERSION string = "2.0.0"
var buildTime string
var gitVersion string

var (
	configFile      = flag.String("config", "dnsmonitor.conf", "Configuration file")
	showVersionFlag = flag.Bool("version", false, "Show dnsconfig version")
	Verbose         = flag.Bool("verbose", false, "verbose output")
)

func init() {
	if len(gitVersion) > 0 {
		VERSION = VERSION + "/" + gitVersion
	}

	log.SetPrefix("geodns ")
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
}

func main() {

	flag.Parse()

	if *showVersionFlag {
		fmt.Println("dnsmonitor", VERSION, buildTime)
		os.Exit(0)
	}

	hub := NewHub()

	cfg, err := configRead(*configFile)
	if err != nil {
		log.Printf("Could not read config file '%s': %s\n", *configFile, err)
		os.Exit(2)
	}

	go startHttp(2090, hub)

	for _, server := range cfg.Servers.A {
		log.Println("Adding", server)
		err := hub.AddName(server)
		if err != nil {
			log.Printf("Could not add '%s': %s\n", server, err)
		}
	}

	// hub.Add("207.171.17.42")
	// hub.Add("199.15.176.152")
	// hub.Add("127.0.0.1")

	hub.AddName("bad-example.develooper.com")

	wg := new(sync.WaitGroup)

	wg.Add(1)
	wg.Wait()
}
