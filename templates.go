package main

import "net/http"

func loadBundle() {
	// nop
}

func serveStatic(w http.ResponseWriter, req *http.Request) {
	fn := http.StripPrefix("/static/", http.FileServer(FS(*devel)))
	fn.ServeHTTP(w, req)
}

func template(name string) ([]byte, error) {
	data, err := FSByte(*devel, "/templates/"+name)
	if err != nil {
		return nil, err
	}
	return data, nil
}
