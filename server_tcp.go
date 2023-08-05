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
	err := server.store.MemPut(ls)
	if err != nil {
		log.Printf("Failed to add RingBuf to cache with error: %v", err)
		return
	}
	defer server.store.Persist(ls.token)

	// Subscribe log state to the TCP buffers being read.
	server.pubsub.Subscribe(ls.token, ls)
	defer server.pubsub.Unsubscribe(ls.token, ls)

	// Write the URL/QR-code on the TCP connection.
	url := fmt.Sprintf("https://%s/view?token=%x\n", PUBLIC_IP, ls.token)
	conn.Write([]byte(url))
	if qr {
		qrterminal.GenerateHalfBlock(url, qrterminal.L, conn)
	}
	log.Printf("[%x] New connection: %s", ls.token, conn.RemoteAddr().String())

	// Read the TCP connection for logs.
	// This funcion does not return until the connection terminates with an EOF.
	server.pollTcpConn(conn, ls.token)
}

func (server *LogoutServer) pollTcpConn(conn net.Conn, token int64) {
	buf := make([]byte, READ_CHUNK_SIZE)
	for {
		readLen, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("[%x] EOF!", token)
				server.pubsub.Publish(token, nil)
			} else {
				log.Printf("[%x] Error reading: %v", token, err.Error())
			}
			break
		}
		log.Printf("[%x] Publishing %d Bytes!", token, readLen)
		// Make a copy to reduce memory footprint and to be able to reuse the read buffer.
		tmp := make([]byte, readLen)
		copy(tmp, buf[:readLen])
		server.pubsub.Publish(token, tmp)
	}
}
