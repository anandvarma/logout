package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (server *LogoutServer) startWebServer(uri string) {
	// Register handlers.
	http.HandleFunc("/", server.staticPageHandler)
	http.HandleFunc("/stream/", server.webSocketHandler)
	http.Handle("/metrics", promhttp.Handler())

	// Block and listen for HTTP requests.
	http.ListenAndServe(uri, nil)
}

func (server *LogoutServer) staticPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Accepted HTTP Connection... %s", r.URL.Path)

	if r.URL.Path == "/" {
		numHttpHome.Inc()
		http.ServeFile(w, r, "./html/index.html")
		return
	}
	if r.URL.Path == "/view" {
		numHttpView.Inc()
		http.ServeFile(w, r, "./html/view.html")
		return
	}

	// 404
	numHttp404.Inc()
	w.Write([]byte("404 Page Not Found!"))
}

func (server *LogoutServer) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Accepted WebSocket Connection... %s", r.URL.Path)
	numHttpStream.Inc()

	streamOp := NewStreamLogOp(r, w, &server.pubsub, &server.store)
	streamOp.Start()
}
