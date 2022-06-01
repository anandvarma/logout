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
	cache     map[int64]*BuffArray
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
		cache:     make(map[int64]*BuffArray, CACHE_SIZE),
	}
}

func (ls *LogServer) Start() {
	defer ls.tcpListen.Close()
	for {
		conn, err := ls.tcpListen.Accept()
		if err != nil {
			log.Fatal("Error accepting: " + err.Error())
		}
		go ls.handleTcpRequest(conn)
	}
}

func (ls *LogServer) StartWeb() {
	http.HandleFunc("/stream/", ls.homePageHandler)
	http.ListenAndServe(":8080", nil)
}

func (ls *LogServer) handleTcpRequest(conn net.Conn) {
	drainOp := NewDrainLogOp(conn)
	defer drainOp.Close()
	drainOp.Greet()
	drainOp.Drain()

	ls.cache[drainOp.tokenNum] = drainOp.ReleaseBuff()
	fmt.Println(ls.cache)
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

	buff, found := ls.cache[n]
	if !found {
		log.Printf("[%x] Cache Miss!", n)
		fmt.Fprintf(w, "Cache Miss!")
		return
	}

	log.Printf("[%x] Cache Hit!", n)

	tmp := new(strings.Builder)
	io.Copy(tmp, buff)

	fmt.Println(tmp.String())
	fmt.Fprint(w, tmp.String())
}
