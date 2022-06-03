package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
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
	http.HandleFunc("/", ls.staticPageHandler)
	http.HandleFunc("/stream", ls.webSocketHandler)
	http.ListenAndServe(":8080", nil)
}

func (ls *LogServer) staticPageHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Accepted HTTP Connection... %s", r.URL.Path)

	if r.URL.Path == "/" {
		http.ServeFile(w, r, "index.html")
		return
	}
	if r.URL.Path == "/view" {
		http.ServeFile(w, r, "view.html")
		return
	}

	// 404
	w.Write([]byte("404 Page Not Found!"))
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (ls *LogServer) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Accepted WebSocket Connection...")
	wsUpgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
	}
	defer ws.Close()

	ws.WriteMessage(1, []byte("Hiiiiii!!"))
	if err != nil {
		log.Printf("WebSocket Write failed: %v", err)
	}
}

func (ls *LogServer) oldie(w http.ResponseWriter, r *http.Request) {
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
