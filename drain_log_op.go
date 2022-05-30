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
	tokenNum int32
}

func NewDrainLogOp(conn net.Conn) *DrainLogOp {
	return &DrainLogOp{
		conn:     conn,
		buffArr:  nil,
		tokenNum: rand.Int31(),
	}
}

func (lConn *DrainLogOp) Drain() {
	log.Printf("[%x] New connection: %s", lConn.tokenNum, lConn.conn.RemoteAddr().String())
	lConn.buffArr = NewBuffArray(BUFF_ARR_CAP)
	for {
		buf := make([]byte, READ_CHUNK_SIZE)
		readLen, err := lConn.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("[%x] EOF!", lConn.tokenNum)
			} else {
				log.Printf("[%x] Error reading: %v", lConn.tokenNum, err.Error())
			}
			break
		}
		log.Printf("[%x] Read %d Bytes!", lConn.tokenNum, readLen)
		lConn.buffArr.AddBuff(buf)
	}
}

func (lConn *DrainLogOp) ReleaseBuff() *BuffArray {
	ret := lConn.buffArr
	lConn.buffArr = nil
	return ret
}

func (lConn *DrainLogOp) Greet() {
	lConn.conn.Write([]byte(fmt.Sprintf("Streaming out logs at: %x", lConn.tokenNum)))
}

func (lConn *DrainLogOp) Close() {
	lConn.conn.Close()
}
