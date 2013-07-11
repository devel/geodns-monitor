package main

import (
	"code.google.com/p/gcfg"
	"log"
	"net"
	"os"
	"strings"
)

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
			err := hub.AddName(strings.Join(nameSlice, "."))
			if err != nil {
				log.Printf("Could not add '%s': %s\n", name, err)
			}
		}
	}

	hub.MarkConfigurationEnd()

}
