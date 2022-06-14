package main

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type LogServer struct {
	uri       string
	tcpListen net.Listener
	pubsub    PubSub
	ls        LogStore
}

func NewLogServer(uri string) *LogServer {
	l, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Error listening on %s", uri)
	}
	log.Printf("Listening on %s ...", uri)

	// Seed rand to avoid collisions after a restart.
	rand.Seed(time.Now().UnixNano())

	return &LogServer{
		uri:       uri,
		tcpListen: l,
		pubsub:    *NewPubSub(),
		ls:        *NewLogStore(),
	}
}

func (ls *LogServer) Start() {
	defer ls.tcpListen.Close()
	for {
		conn, err := ls.tcpListen.Accept()
		if err != nil {
			log.Fatal("Error accepting: " + err.Error())
		}

		drainOp := NewDrainLogOp(conn, &ls.pubsub, &ls.ls)
		go drainOp.Start()
	}
}

func (ls *LogServer) StartWeb() {
	http.HandleFunc("/", ls.staticPageHandler)
	http.HandleFunc("/stream/", ls.webSocketHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(HOST+":"+WEB_PORT, nil)
}

func (ls *LogServer) staticPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Accepted HTTP Connection... %s", r.URL.Path)

	if r.URL.Path == "/" {
		http.ServeFile(w, r, "./html/index.html")
		return
	}
	if r.URL.Path == "/view" {
		http.ServeFile(w, r, "./html/view.html")
		return
	}

	// 404
	w.Write([]byte("404 Page Not Found!"))
}

func (ls *LogServer) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Accepted WebSocket Connection... %s", r.URL.Path)
	streamOp := NewStreamLogOp(r, w, &ls.pubsub, &ls.ls)
	streamOp.Start()
}
