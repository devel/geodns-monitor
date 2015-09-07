package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gorilla/mux"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	templateFile, err := template("index.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	w.Write(templateFile)
}

func statusHandler(hub *StatusHub) func(rest.ResponseWriter, *rest.Request) {
	return func(w rest.ResponseWriter, _ *rest.Request) {

		currentStatus := hub.Status()

		type apiStatus struct {
			Status
			LastUpdatedAgo string `json:"last_update"`
			Restarted      string `json:"uptime_p"`
		}

		byIP := make(map[string]*apiStatus)

		for _, st := range currentStatus {

			var lastUpdatedAgoStr, uptimeStr string

			lastUpdatedAgo := DayDuration{time.Since(st.LastStatusUpdate)}
			uptime := DayDuration{time.Since(time.Unix(time.Now().Unix()-st.Uptime, 0))}

			if uptime.Seconds() <= lastUpdatedAgo.Seconds() {
				uptimeStr = ""
			} else {
				uptimeStr = uptime.DayString()
			}

			if lastUpdatedAgo.Seconds() > 1 {
				lastUpdatedAgoStr = lastUpdatedAgo.DayString()
			} else {
				lastUpdatedAgoStr = "now"
			}

			rv := &apiStatus{
				*st,
				lastUpdatedAgoStr,
				uptimeStr,
			}

			byIP[st.IP] = rv
		}

		// remoteIP := req.RemoteAddr

		w.WriteJson(map[string]interface{}{"servers": byIP})
	}
}

func setupMux(hub *StatusHub) http.Handler {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	apirouter, err := rest.MakeRouter(
		rest.Get("/status", statusHandler(hub)),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(apirouter)

	router := mux.NewRouter()
	router.HandleFunc("/", homeHandler)
	router.PathPrefix("/api/").Handler(http.StripPrefix("/api", api.MakeHandler()))
	router.PathPrefix("/static/").HandlerFunc(serveStatic)

	smux := http.NewServeMux()
	smux.Handle("/", router)

	return smux
}

func startHTTP(port int, hub *StatusHub) {
	h := setupMux(hub)

	listen := ":" + strconv.Itoa(port)
	log.Println("Going to listen on port", listen)

	// handlers.CombinedLoggingHandler(os.Stdout,
	http.ListenAndServe(listen, h)
}
