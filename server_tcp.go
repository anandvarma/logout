package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/mdp/qrterminal"
)

func (server *LogoutServer) startTcpServer(uri string) {
	tcpListen, err := net.Listen("tcp", uri)
	if err != nil {
		log.Fatalf("Error listening on %s", uri)
	}
	log.Printf("Listening on %s ...", uri)
	defer tcpListen.Close()

	for {
		conn, err := tcpListen.Accept()
		if err != nil {
			log.Fatal("Error accepting: " + err.Error())
		}
		numTcpAccepts.Inc()
		go server.handleTcpRequest(conn, PRINT_QR_CODE)
	}
}

func (server *LogoutServer) handleTcpRequest(conn net.Conn, qr bool) {
	// Close the TCP connection on the way out.
	defer conn.Close()

	// Set up memory state for this log and persist when done.
	ls := newLogState()
	err := server.manager.OpenLog(ls.logKey, &ls.logVal)
	if err != nil {
		log.Printf("Failed to add logState to cache with error: %v", err)
		return
	}
	defer server.manager.CloseLog(ls.logKey)

	// Subscribe log state to the TCP buffers being read.
	server.pubsub.Subscribe(ls.logKey, ls)
	defer server.pubsub.Unsubscribe(ls.logKey, ls)

	// Write the URL/QR-code on the TCP connection.
	url := fmt.Sprintf("https://%s/view?token=%x\n", PUBLIC_IP, ls.token)
	conn.Write([]byte(url))
	if qr {
		qrterminal.GenerateHalfBlock(url, qrterminal.L, conn)
	}
	log.Printf("[%v] New connection: %s", ls.logKey, conn.RemoteAddr().String())

	// Read the TCP connection for logs.
	// This funcion does not return until the connection terminates with an EOF.
	server.pollTcpConn(conn, ls.logKey)
}

func (server *LogoutServer) pollTcpConn(conn net.Conn, key logKey) {
	buf := make([]byte, READ_CHUNK_SIZE)
	for {
		readLen, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("[%v] EOF!", key)
				server.pubsub.Publish(key, nil)
			} else {
				log.Printf("[%v] Error reading: %v", key, err.Error())
			}
			break
		}
		log.Printf("[%v] Publishing %d Bytes!", key, readLen)
		// Make a copy to reduce memory footprint and to be able to reuse the read buffer.
		tmp := make([]byte, readLen)
		copy(tmp, buf[:readLen])
		server.pubsub.Publish(key, tmp)
	}
}
