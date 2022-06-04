package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type StreamLogOp struct {
	r      *http.Request
	w      http.ResponseWriter
	pubsub *PubSub
	cache  *LogCache
	ws     *websocket.Conn
	lock   sync.Mutex // serialize access to the websocket
}

func NewStreamLogOp(r *http.Request, w http.ResponseWriter, pubsub *PubSub, cache *LogCache) *StreamLogOp {
	return &StreamLogOp{
		r:      r,
		w:      w,
		pubsub: pubsub,
		cache:  cache,
		ws:     nil,
	}
}

func (op *StreamLogOp) Start() {
	defer op.Finish()

	// Parse stream identifier token.
	res := strings.TrimPrefix(op.r.URL.Path, "/stream/")
	if len(res) == 0 {
		fmt.Fprintf(op.w, "Invalid stream token!")
		return
	}
	n, err := strconv.ParseInt(res, 16, 64)
	if err != nil {
		fmt.Fprintf(op.w, "Invalid stream token: %s", res)
		return
	}
	log.Printf("[%x] %v %v", n, op.r.Method, op.r.RemoteAddr)

	// Upgrade to websocket.
	var wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	wsUpgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := wsUpgrader.Upgrade(op.w, op.r, nil)
	if err != nil {
		log.Printf("Upgrade failed: %v", err)
		return
	}
	op.ws = ws

	op.GetLogsForToken(n)
}

func (op *StreamLogOp) GetLogsForToken(token int64) {
	buff, found := op.cache.Get(token)
	if !found {
		log.Printf("[%x] Cache Miss!", token)
		op.ws.WriteMessage(websocket.TextMessage, []byte("No logs found!"))
		return
	}

	if buff.IsClosed() {
		// Log stream closed. Send the ring buffer contents out.
		log.Printf("[%x] Cache Hit!", token)
		tmp := new(strings.Builder)
		io.Copy(tmp, buff)
		fmt.Println(tmp.String())
		op.ws.WriteMessage(websocket.TextMessage, []byte(tmp.String()))
		return
	}

	// Write in progress, become a subscriber.
	op.lock.Lock()
	log.Printf("[%x] Streaming...", token)
	op.pubsub.Subscribe(token, op)
	op.lock.Unlock()

	// Block until the socket is open and read.
	op.ListenForMessages()
}

func (op *StreamLogOp) ListenForMessages() {
	for op.ws != nil {
		mType, msg, err := op.ws.ReadMessage()
		if err != nil {
			log.Printf("Failed to read from socket: %v", err)
			break
		}
		log.Printf("Socket read: %d %s %v", mType, string(msg), err)
	}

	op.Finish()
}

func (op *StreamLogOp) SubCb(buf []byte) {
	// SubCb runs on the publisher's thread. Spin a go routine to avoid blocking.
	go op.WriteLog(buf)
}

func (op *StreamLogOp) WriteLog(buf []byte) {
	op.lock.Lock()
	defer op.lock.Unlock()

	err := op.ws.WriteMessage(websocket.TextMessage, buf)
	log.Printf("STREAM >> SubCb() : %v", err)
}

func (op *StreamLogOp) Finish() {
	op.ws.Close()
}
