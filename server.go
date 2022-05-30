package main

import (
	"fmt"
	"log"
	"net"
)

type LogServer struct {
	uri       string
	tcpListen net.Listener
	cache     map[int32]*BuffArray
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
		cache:     make(map[int32]*BuffArray, CACHE_SIZE),
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

func (ls *LogServer) handleTcpRequest(conn net.Conn) {
	drainOp := NewDrainLogOp(conn)
	defer drainOp.Close()
	drainOp.Greet()
	drainOp.Drain()

	ls.cache[drainOp.tokenNum] = drainOp.ReleaseBuff()
	fmt.Println(ls.cache)
}
