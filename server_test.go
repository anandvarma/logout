package main

import (
	"log"
	"net"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLoogutServer(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logout := NewLogoutServer()

	// Setup a test pipe.
	server, client := net.Pipe()
	go logout.handleTcpRequest(server, false)

	// Check to make sure the server sends back the URI.
	uri := readLine(client)
	token, err := parseToken(uri)
	checkEq(t, err, nil)

	// Sanity check internal structures.
	checkEq(t, len(logout.store.cache), 1)
	ls, exists := logout.store.MemGet(token)
	checkEq(t, exists, true)
	checkEq(t, ls.IsActive(), true)
	checkEq(t, ls.token, token)
	checkEq(t, ls.rb.Len(), 0)

	// Write some logs on the client.
	client.Write([]byte("Foo"))
	client.Write([]byte("Bar"))
	// Validate logs on the server.
	time.Sleep(100 * time.Millisecond)
	checkEq(t, ls.rb.Len(), 2)
	checkEq(t, ls.rb.arr[0], []byte("Foo"))
	checkEq(t, ls.rb.arr[1], []byte("Bar"))

	// Close the client connection and validate server state.
	client.Close()
	time.Sleep(100 * time.Millisecond)
	checkEq(t, ls.IsActive(), false)
	_, exists = logout.store.MemGet(token)
	checkEq(t, exists, false)
	checkEq(t, len(logout.store.cache), 0)
}

func readLine(conn net.Conn) string {
	buf := make([]byte, 1024*1024)
	readLen, _ := conn.Read(buf)
	return strings.TrimSuffix(string(buf[:readLen]), "\n")
}

func parseToken(uri string) (int64, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(u.Query()["token"][0], 16, 64)
	return n, err
}
