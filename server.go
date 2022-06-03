package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type LogServer struct {
	uri       string
	tcpListen net.Listener
	pubsub    PubSub
	cache     LogCache
}

func NewLogServer(uri string) *LogServer {
	l, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Error listening on %s", uri)
	}
	log.Printf("Listening on %s ...", uri)

	return &LogServer{
		uri:       uri,
		tcpListen: l,
		pubsub:    *NewPubSub(),
		cache:     *NewLogCache(),
	}
}

func (ls *LogServer) Start() {
	defer ls.tcpListen.Close()
	for {
		conn, err := ls.tcpListen.Accept()
		if err != nil {
			log.Fatal("Error accepting: " + err.Error())
		}

		drainOp := NewDrainLogOp(conn, &ls.pubsub, &ls.cache)
		go drainOp.Start()
	}
}

func (ls *LogServer) StartWeb() {
	http.HandleFunc("/stream/", ls.homePageHandler)
	http.ListenAndServe(":8080", nil)
}

func (ls *LogServer) homePageHandler(w http.ResponseWriter, r *http.Request) {
	// Trim prefix and prase stream id.
	res := strings.TrimPrefix(r.URL.Path, "/stream/")
	if len(res) == 0 {
		fmt.Fprintf(w, "Invalid stream ID!")
		return
	}
	n, err := strconv.ParseInt(res, 16, 64)
	if err != nil {
		fmt.Fprintf(w, "Invalid stream ID!")
		return
	}

	log.Printf("[%x] %v %v", n, r.Method, r.RemoteAddr)

	buff, found := ls.cache.Get(n)
	if !found {
		log.Printf("[%x] Cache Miss!", n)
		fmt.Fprintf(w, "Cache Miss!")
		return
	}

	if !buff.IsClosed() {
		// Write in progress, become a subscriber.
		log.Printf("[%x] Streaming...", n)
		streamOp := NewStreamLogOp(w)
		streamOp.Start()
		ls.pubsub.Subscribe(n, streamOp)
		return
	}

	// Log stream closed. Send the ring buffer contents out.
	log.Printf("[%x] Cache Hit!", n)
	tmp := new(strings.Builder)
	io.Copy(tmp, buff)
	fmt.Println(tmp.String())
	fmt.Fprint(w, tmp.String())
}
