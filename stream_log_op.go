package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

type StreamLogOp struct {
	r      *http.Request
	w      http.ResponseWriter
	pubsub *PubSub
	cache  *LogCache
	ws     *websocket.Conn
	token  int64
	queue  chan []byte
}

func NewStreamLogOp(r *http.Request, w http.ResponseWriter, pubsub *PubSub, cache *LogCache) *StreamLogOp {
	return &StreamLogOp{
		r:      r,
		w:      w,
		pubsub: pubsub,
		cache:  cache,
		ws:     nil,
		token:  0,
		queue:  nil,
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

	// Fetch the in memory stats.
	isLive := op.GetMemLogs()
	if !isLive {
		return
	}

	// Write in progress, become a subscriber and spawn a writer to consume incoming logs.
	log.Printf("[%x] Streaming...", op.token)
	op.queue = make(chan []byte, STREAM_QUEUE_SIZE)
	defer close(op.queue)
	defer log.Printf("[%x] close(op.queue)", op.token)

	op.pubsub.Subscribe(op.token, op)
	defer op.pubsub.Unsubscribe(op.token, op)
	defer log.Printf("[%x] pubsub.Unsubscribe()", op.token)

	// Start a stream writer in the background to do the work.
	go op.streamWriter()

	// Read Loop. Blocks until websocket gets terminated by the client.
	for {
		mt, b, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Failed to read from socket: %v", err)
			break
		}
		log.Printf("Socket read: %d %s %v", mt, string(b), err)
	}
	log.Printf("[%x] Start() Completed!", op.token)
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

func (op *StreamLogOp) GetMemLogs() bool {
	buff, found := op.cache.Get(op.token)
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

func (op *StreamLogOp) streamWriter() {
	defer log.Printf("[%x] streamWriter goroutine ending...", op.token)

	for buf := range op.queue {
		fmt.Println("Dequeued sucessfully")
		err := op.ws.WriteMessage(websocket.TextMessage, buf)
		log.Printf("STREAM >> : %v", err)
		if err != nil {
			log.Println("Write to websocket failed, stopping stream!")
			return
		}
	}
}

func (op *StreamLogOp) SubCb(buf []byte) bool {
	// This function runs on the publisher's thread. Avoid blocking.
	select {
	case op.queue <- buf:
		fmt.Println("Enqueued sucessfully.")
		return true
	default:
		// Queue full.
		// The websocket is unable to keep up with the flow of incoming logs.
		// Unsubscribe from further events.
		return false
	}
}
