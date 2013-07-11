// +build !devel

package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func loadBundle() {
	for k, v := range _bundle {

		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			log.Println(err)
		} else {
			_bundle[k] = string(b)
		}
	}
}

func serveStatic(w http.ResponseWriter, req *http.Request) {
	f, ok := _bundle[strings.TrimLeft(req.URL.Path, "/")]
	if ok {
		w.Write([]byte(f))
	} else {
		http.NotFound(w, req)
	}
}

func template(name string) ([]byte, error) {
	data, ok := _bundle["templates/"+name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s", name)
	}
	return []byte(data), nil
}

func status_html() []byte {
	data, err := ioutil.ReadFile("status.html")
	if err != nil {
		log.Println("Could not open status.html", err)
		return []byte{}
	}
	return data
}
