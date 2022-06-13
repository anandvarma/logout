package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"

	"github.com/mdp/qrterminal"
)

type DrainLogOp struct {
	tokenNum int64
	rb       *RingBuf
	conn     net.Conn
	pubsub   *PubSub
	ls       *LogStore
}

func NewDrainLogOp(conn net.Conn, pubsub *PubSub, ls *LogStore) *DrainLogOp {
	return &DrainLogOp{
		tokenNum: rand.Int63n(MAX_TOKEN_NUM),
		rb:       NewRingBuf(BUFF_ARR_CAP),
		conn:     conn,
		pubsub:   pubsub,
		ls:       ls,
	}
}

func (op *DrainLogOp) Start() {
	defer op.Finish()

	// Greet.
	url := fmt.Sprintf("http://%s:%s/view?token=%x\n", HOST, WEB_PORT, op.tokenNum)
	op.conn.Write([]byte(url))
	if PRINT_QR_CODE {
		qrterminal.GenerateHalfBlock(url, qrterminal.L, op.conn)
	}
	log.Printf("[%x] New connection: %s", op.tokenNum, op.conn.RemoteAddr().String())

	// Add the open RingBuf to cache.
	err := op.ls.MemPut(op.tokenNum, op.rb)

	if err != nil {
		log.Printf("Failed to add RingBuf to cache with error: %v", err)
		return
	}

	// Initialize in memory subscriber and publish TCP read buffers.
	op.pubsub.Subscribe(op.tokenNum, op.rb)
	op.PublishLoop()
}

func (op *DrainLogOp) PublishLoop() {
	buf := make([]byte, READ_CHUNK_SIZE)
	for {
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
		// Make a copy to reduce memory footprint and to be able to reuse the read buffer.
		tmp := make([]byte, readLen)
		copy(tmp, buf[:readLen])
		op.pubsub.Publish(op.tokenNum, tmp)
	}
}

func (op *DrainLogOp) Finish() {
	log.Printf("[%x] Finish", op.tokenNum)
	op.conn.Close()

	// Flush to persistent storage.
	op.ls.Persist(op.tokenNum)
}

func (op *DrainLogOp) Token() int64 {
	return op.tokenNum
}
