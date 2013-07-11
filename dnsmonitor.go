package main

import (
	"flag"
	"fmt"
	"log"
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

	loadBundle()

	hub := NewHub()

	go startHttp(2090, hub)

	configure(hub)

	quit := make(chan bool)
	<-quit
}
