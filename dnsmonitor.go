package main

//go:generate esc -o static.go -ignore .DS_Store -prefix static templates static

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

// VERSION is the application version number
var VERSION = "2.1.0"
var buildTime string
var gitVersion string

var (
	configFile      = flag.String("config", "dnsmonitor.conf", "Configuration file")
	showVersionFlag = flag.Bool("version", false, "Show dnsconfig version")
	verbose         = flag.Bool("verbose", false, "verbose output")
	devel           = flag.Bool("devel", false, "Use development assets")
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

	go startHTTP(2090, hub)

	go func() {
		for {
			log.Println("running configuration...")
			configure(hub)
			time.Sleep(20 * time.Second)
		}
	}()

	quit := make(chan bool)
	<-quit
}
