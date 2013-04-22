package main

import (
	"code.google.com/p/gcfg"
	"log"
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
