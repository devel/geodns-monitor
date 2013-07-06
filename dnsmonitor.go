package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
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

	go startHttp(2090, hub)

	configure(hub)

	quit := make(chan bool)
	<-quit
}

func configure(hub *StatusHub) {

	cfg, err := configRead(*configFile)
	if err != nil {
		log.Printf("Could not read config file '%s': %s\n", *configFile, err)
		os.Exit(2)
	}

	hub.MarkConfigurationStart()

	for _, server := range cfg.Servers.A {
		log.Println("Adding", server)
		err := hub.AddName(server)
		if err != nil {
			log.Printf("Could not add '%s': %s\n", server, err)
		}
	}

	for _, domain := range cfg.Servers.Domain {
		log.Println("Adding NSes for", domain)

		nses, err := net.LookupNS(domain)

		log.Printf("NSes: %#v: %s\n", nses, err)
		for _, ns := range nses {
			log.Printf("Adding '%s'\n", ns.Host)
			err := hub.AddName(ns.Host)
			if err != nil {
				log.Printf("Could not add '%s': %s\n", ns.Host, err)
			}
		}
	}

	hub.MarkConfigurationEnd()

}
