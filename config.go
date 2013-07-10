package main

import (
	"code.google.com/p/gcfg"
	"log"
	"net"
	"os"
)

type AppConfig struct {
	Servers struct {
		A  []string
		Ns []string
	}
}

func configRead(fileName string) (*AppConfig, error) {
	cfg := new(AppConfig)

	err := gcfg.ReadFileInto(cfg, fileName)
	if err != nil {
		log.Printf("Failed to parse config data: %s\n", err)
		return nil, err
	}
	return cfg, nil
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
