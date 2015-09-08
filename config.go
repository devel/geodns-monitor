package main

import (
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"code.google.com/p/gcfg"
)

// AppConfig is the 'master' application configuration
type AppConfig struct {
	Servers struct {
		A      []string
		Domain []string
		Txt    []string
	}
}

func configRead(fileName string) (*AppConfig, error) {
	cfg := new(AppConfig)

	err := gcfg.ReadFileInto(cfg, fileName)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func configure(hub *StatusHub) {

	cfg, err := configRead(*configFile)
	if err != nil {
		log.Printf("Could not read config file: %s\n", err)
		os.Exit(2)
	}

	hub.MarkConfigurationStart()
	wg := &sync.WaitGroup{}
	errch := make(chan error, 20)

	for _, server := range cfg.Servers.A {
		log.Println("Adding", server)
		wg.Add(1)
		hub.AddNameBackground(server, errch)
	}

	for _, domain := range cfg.Servers.Domain {
		log.Println("Adding NSes for", domain)

		nses, err := net.LookupNS(domain)

		log.Printf("NSes: %#v: %s\n", nses, err)
		for _, ns := range nses {
			log.Printf("Adding '%s'\n", ns.Host)
			wg.Add(1)
			hub.AddNameBackground(ns.Host, errch)
		}
	}

	for _, txtconfig := range cfg.Servers.Txt {

		x := strings.SplitN(txtconfig, ",", 2)
		txtname := strings.TrimSpace(x[0])
		txtbase := strings.TrimSpace(x[1])

		log.Println("Adding TXT for", txtname, txtbase)

		txts, err := net.LookupTXT(txtname)
		log.Printf("TXTs: %#v: %s\n", txts, err)

		names := []string{}

		for _, txt := range txts {
			for _, name := range strings.Split(txt, " ") {
				names = append(names, name)
			}
		}

		nameSlice := []string{"", txtbase}
		for _, name := range names {
			nameSlice[0] = name
			hub.AddNameBackground(strings.Join(nameSlice, "."), errch)
		}
	}

	go func() {
		for err := range errch {
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}
	}()

	wg.Wait()
	close(errch)
	hub.MarkConfigurationEnd()

}
