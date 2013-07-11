// +build devel

package main

import (
	"io/ioutil"
	// "log"
	"net/http"
)

func loadBundle() {
	// nop
}

// http.Handle("/static/", ))

func serveStatic(w http.ResponseWriter, req *http.Request) {
	fn := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	fn.ServeHTTP(w, req)
}

func template(name string) ([]byte, error) {
	data, err := ioutil.ReadFile("templates/" + name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// func serveStatic() []byte {
// 	data, err := ioutil.ReadFile("status.html")
// 	if err != nil {
// 		log.Println("Could not open status.html", err)
// 		return []byte{}
// 	}
// 	return data
// }
