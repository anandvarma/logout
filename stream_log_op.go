package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type StreamLogOp struct {
	r        *http.Request
	w        http.ResponseWriter
	pubsub   *PubSub
	cache    *LogStore
	ws       *websocket.Conn
	token    int64
	queue    chan []byte
	ctx      context.Context
	doneChan chan struct{}
}

func NewStreamLogOp(r *http.Request, w http.ResponseWriter, pubsub *PubSub, cache *LogStore) *StreamLogOp {
	return &StreamLogOp{
		r:      r,
		w:      w,
		pubsub: pubsub,
		cache:  cache,
		ws:     nil,
		token:  0,
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
	defer log.Printf("[%x] ws.Close()", op.token)

	// Parse stream identification token.
	ok := op.ParseToken()
	if !ok {
		op.ws.WriteMessage(websocket.TextMessage, []byte("Bad stream token!"))
		fmt.Println("404")
		return
	}

	// Fetch persisted logs.
	done := op.GetPersistentLogs()
	if done {
		return
	}

	// Fetch the in memory logs.
	isLive := op.GetMemLogs()
	if !isLive {
		return
	}
	log.Printf("[%x] Streaming...", op.token)
	op.StreamLogs()
}

func (op *StreamLogOp) ParseToken() bool {
	res := strings.TrimPrefix(op.r.URL.Path, "/stream/")
	if len(res) == 0 {
		fmt.Println("Invalid stream token!")
		return false
	}
	n, err := strconv.ParseInt(res, 16, 64)
	if err != nil {
		fmt.Printf("Invalid stream token: %s", res)
		return false
	}
	op.token = n
	log.Printf("[%x] %v %v", op.token, op.r.Method, op.r.RemoteAddr)
	return true
}

func (op *StreamLogOp) GetPersistentLogs() bool {
	buf := op.cache.db.GetLogBuf(op.token)
	if buf == nil {
		return false
	}
	op.ws.WriteMessage(websocket.TextMessage, buf)
	return true
}

func (op *StreamLogOp) GetMemLogs() bool {
	buff, found := op.cache.MemGet(op.token)
	if !found {
		log.Printf("[%x] Cache Miss!", op.token)
		op.ws.WriteMessage(websocket.TextMessage, []byte("No logs found!"))
		return false
	}

	// Read the ring buffer and add to the streaming queue.
	log.Printf("[%x] Cache Hit!", op.token)
	tmp := new(strings.Builder)
	io.Copy(tmp, buff)
	op.ws.WriteMessage(websocket.TextMessage, []byte(tmp.String()))
	fmt.Println(tmp.String())
	return !buff.IsClosed()
}

func (op *StreamLogOp) StreamLogs() {
	ctx, cancel := context.WithCancel(context.Background())
	op.ctx = ctx
	op.doneChan = make(chan struct{}, 3)
	op.queue = make(chan []byte, STREAM_QUEUE_SIZE)

	// Subscribe to live logs and produce into the buffer.
	op.pubsub.Subscribe(op.token, op)
	// Start a stream writer in the background to consume buffered logs.
	go op.streamWriter()
	// Read the websocket for close and other control messages.
	go op.socketReader()

	// Stop if we get a channel notification from any one of socketReader,
	// streamWriter or SubCb.
	<-op.doneChan
	cancel()

	log.Printf("[%x] Done, cleaning up...", op.token)
	op.pubsub.Unsubscribe(op.token, op)
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
	defer log.Printf("[%x] streamWriter goroutine ending...", op.token)

	for {
		select {
		case <-op.ctx.Done():
			return
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
	defer log.Printf("[%x] socketReader goroutine ending...", op.token)

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
