package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
)

type DrainLogOp struct {
	conn     net.Conn
	buffArr  *BuffArray
	tokenNum int64
}

func NewDrainLogOp(conn net.Conn) *DrainLogOp {
	return &DrainLogOp{
		conn:     conn,
		buffArr:  nil,
		tokenNum: rand.Int63n(100),
	}
}

func (op *DrainLogOp) Drain() {
	log.Printf("[%x] New connection: %s", op.tokenNum, op.conn.RemoteAddr().String())
	op.buffArr = NewBuffArray(BUFF_ARR_CAP)
	for {
		buf := make([]byte, READ_CHUNK_SIZE)
		readLen, err := op.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("[%x] EOF!", op.tokenNum)
			} else {
				log.Printf("[%x] Error reading: %v", op.tokenNum, err.Error())
			}
			break
		}
		log.Printf("[%x] Read %d Bytes!", op.tokenNum, readLen)
		op.buffArr.AddBuff(buf)
	}
}

func (op *DrainLogOp) ReleaseBuff() *BuffArray {
	ret := op.buffArr
	op.buffArr = nil
	return ret
}

func (op *DrainLogOp) Greet() {
	op.conn.Write([]byte(fmt.Sprintf("Streaming out logs at: %x", op.tokenNum)))
}

func (op *DrainLogOp) Close() {
	op.conn.Close()
}
