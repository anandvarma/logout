package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
)

type DrainLogOp struct {
	tokenNum int64
	rbuf     *RingBuf
	conn     net.Conn
	pubsub   *PubSub
	cache    *LogCache
}

func NewDrainLogOp(conn net.Conn, pubsub *PubSub, cache *LogCache) *DrainLogOp {
	return &DrainLogOp{
		tokenNum: rand.Int63n(100),
		rbuf:     NewRingBuf(BUFF_ARR_CAP),
		conn:     conn,
		pubsub:   pubsub,
		cache:    cache,
	}
}

func (op *DrainLogOp) Start() {
	defer op.Finish()

	// Greet.
	op.conn.Write([]byte(fmt.Sprintf("Streaming out logs at: %x", op.tokenNum)))
	log.Printf("[%x] New connection: %s", op.tokenNum, op.conn.RemoteAddr().String())

	// Add the open RingBuf to cache.
	err := op.cache.Put(op.tokenNum, op.rbuf)

	if err != nil {
		log.Printf("Failed to add RingBuf to cache with error: %v", err)
		return
	}

	// Initialize in memory subscriber and publish TCP read buffers.
	op.pubsub.Subscribe(op.tokenNum, op.rbuf)
	op.PublishLoop()
}

func (op *DrainLogOp) PublishLoop() {
	for {
		buf := make([]byte, READ_CHUNK_SIZE)
		readLen, err := op.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("[%x] EOF!", op.tokenNum)
				op.pubsub.Publish(op.tokenNum, nil)
			} else {
				log.Printf("[%x] Error reading: %v", op.tokenNum, err.Error())
			}
			break
		}
		log.Printf("[%x] Publishing %d Bytes!", op.tokenNum, readLen)
		op.pubsub.Publish(op.tokenNum, buf)
	}
}

func (op *DrainLogOp) Finish() {
	log.Printf("[%x] Finish", op.tokenNum)
	op.conn.Close()
}

func (op *DrainLogOp) Token() int64 {
	return op.tokenNum
}
