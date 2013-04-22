package main

import (
	"github.com/ant0ine/go-json-rest"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	templateFile, err := ioutil.ReadFile("templates/index.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(templateFile)
}

func StatusHandler(hub *StatusHub) func(*rest.ResponseWriter, *rest.Request) {
	return func(w *rest.ResponseWriter, _ *rest.Request) {

		currentStatus := hub.Status()

		type apiStatus struct {
			Status
			LastUpdatedAgo string `json:"last_update"`
			Restarted      string `json:"uptime_p"`
		}

		byIp := make(map[string]*apiStatus)

		for _, st := range currentStatus {
			rv := &apiStatus{
				*st,
				time.Since(st.LastStatusUpdate).String(),
				time.Since(time.Unix(time.Now().Unix()-st.Uptime, 0)).String(),
			}

			byIp[st.Ip] = rv
		}

		// remoteIp := req.RemoteAddr

		w.WriteJson(map[string]interface{}{"servers": byIp})
	}
}

func startHttp(port int, hub *StatusHub) {

	router := mux.NewRouter()
	router.HandleFunc("/", HomeHandler)

	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.Handle("/", http.FileServer(http.Dir(".")))

	restHandler := rest.ResourceHandler{}

	restHandler.SetRoutes(
		rest.Route{"GET", "/api/status", StatusHandler(hub)},
	)

	restHandler.EnableGzip = true
	restHandler.EnableLogAsJson = true
	restHandler.EnableResponseStackTrace = true
	restHandler.EnableStatusService = true

	http.Handle("/api/", &restHandler)

	listen := ":" + strconv.Itoa(port)
	log.Println("Going to listen on port", listen)

	// handlers.CombinedLoggingHandler(os.Stdout,
	http.ListenAndServe(listen, http.DefaultServeMux)
}
