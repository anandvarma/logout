package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type StreamLogOp struct {
	r        *http.Request
	w        http.ResponseWriter
	ws       *websocket.Conn
	server   *LogoutServer
	key      logKey
	queue    chan []byte
	ctx      context.Context
	doneChan chan struct{}
}

func NewStreamLogOp(r *http.Request, w http.ResponseWriter, server *LogoutServer) *StreamLogOp {
	return &StreamLogOp{
		r:      r,
		w:      w,
		ws:     nil,
		server: server,
		key:    logKey{},
		queue:  nil,
		ctx:    nil,
	}
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (op *StreamLogOp) Start() {
	// Upgrade connection to a web socket.
	ws, err := wsUpgrader.Upgrade(op.w, op.r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}
	op.ws = ws
	defer op.ws.Close()

	// Parse stream identification token.
	if err := op.parseKey(); err != nil {
		op.ws.WriteMessage(websocket.TextMessage, []byte("Bad stream token!"))
		return
	}

	// Check if the log being streamed is open and stream from memory.
	r, err := op.server.manager.GetOpenLogReader(op.key)
	if err == nil {
		if err := op.copyToWebSocket(r); err != nil {
			return
		}
		op.streamLogs()
		return
	}

	// Check logStore for persisted logs that have already been closed.
	r, err = op.server.manager.ls.GetClosedLogReader(op.key)
	if err == nil {
		op.copyToWebSocket(r)
		return
	}

	// Log not found in memory or in the log store.
	op.ws.WriteMessage(websocket.TextMessage, []byte("No logs found!"))
}

func (op *StreamLogOp) parseKey() error {
	res := strings.TrimPrefix(op.r.URL.Path, "/stream/")
	if len(res) == 0 {
		log.Println("Invalid stream token!")
		return ErrInvalidKey
	}
	n, err := strconv.ParseInt(res, 16, 64)
	if err != nil {
		log.Printf("Invalid stream token: %s", res)
		return ErrInvalidKey
	}
	op.key = logKey{owner: ANONYMOUS_OWNER_ID, token: n}
	log.Printf("[%v] %v %v", op.key, op.r.Method, op.r.RemoteAddr)
	return nil
}

func (op *StreamLogOp) copyToWebSocket(r io.Reader) error {
	w, err := op.ws.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Printf("[%v] Failed to get WebSocket Writer with %v", op.key, err)
		return err
	}

	if _, err := io.Copy(w, r); err != nil {
		log.Printf("[%v] Failed to Copy with %v", op.key, err)
		return err
	}

	if err := w.Close(); err != nil {
		log.Printf("[%v] Failed to Close with %v", op.key, err)
		return err
	}

	return nil
}

func (op *StreamLogOp) streamLogs() {
	log.Printf("[%v] Streaming...", op.key)
	ctx, cancel := context.WithCancel(context.Background())
	op.ctx = ctx
	op.doneChan = make(chan struct{}, 3)
	op.queue = make(chan []byte, STREAM_QUEUE_SIZE)

	// Subscribe to live logs and produce into the buffer.
	op.server.pubsub.Subscribe(op.key, op)
	// Start a stream writer in the background to consume buffered logs.
	go op.streamWriter()
	// Read the websocket for close and other control messages.
	go op.socketReader()

	// Stop if we get a channel notification from any one of socketReader,
	// streamWriter or SubCb.
	<-op.doneChan
	cancel()

	log.Printf("[%v] Done, cleaning up...", op.key)
	op.server.pubsub.Unsubscribe(op.key, op)
	close(op.queue)
}

func (op *StreamLogOp) SubCb(buf []byte) bool {
	// This function runs on the publisher's thread. Avoid blocking.
	select {
	case <-op.ctx.Done():
		return false
	case op.queue <- buf:
		fmt.Println("Enqueued successfully.")
		return true // Stay subscribed
	default:
		log.Println("Stream buffer full, unable to keep up with incoming logs")
		op.doneChan <- struct{}{}
		return false
	}
}

func (op *StreamLogOp) streamWriter() {
	defer log.Printf("[%v] streamWriter goroutine ending...", op.key)

	hbTicker := time.NewTicker(30 * time.Second)
	defer hbTicker.Stop()

	for {
		select {
		case <-op.ctx.Done():
			return
		case <-hbTicker.C:
			log.Println("Sending Heartbeat")
			err := op.ws.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				log.Println("Ping to websocket failed, stopping stream!")
				op.doneChan <- struct{}{}
				return
			}
		case buf := <-op.queue:
			if buf == nil {
				fmt.Println("Got EOF")
				op.doneChan <- struct{}{}
				return
			}
			err := op.ws.WriteMessage(websocket.TextMessage, buf)
			log.Printf("STREAM >> : %v", err)
			if err != nil {
				log.Println("Write to websocket failed, stopping stream!")
				op.doneChan <- struct{}{}
				return
			}
		}
	}
}

func (op *StreamLogOp) socketReader() {
	defer log.Printf("[%v] socketReader goroutine ending...", op.key)

	// Read Loop. Blocks until websocket gets terminated by the client.
	for {
		if op.ctx.Err() != nil {
			return
		}

		mt, b, err := op.ws.ReadMessage()
		if err != nil {
			log.Printf("Failed to read from socket: %v", err)
			op.doneChan <- struct{}{}
			return
		}
		log.Printf("Socket read: %d %s %v", mt, string(b), err)
	}
}
