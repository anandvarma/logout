package main

import (
	"log"
	"net/http"
)

type StreamLogOp struct {
	out http.ResponseWriter
}

func NewStreamLogOp(out http.ResponseWriter) *StreamLogOp {
	return &StreamLogOp{
		out: out,
	}
}

func (op *StreamLogOp) Start() {
	// Upgrade to websocket.
	op.out.Header().Set("Content-Type", "text/plain")
}

func (op *StreamLogOp) SubCb(buf []byte) {
	n, err := op.out.Write(buf)
	log.Printf("STREAM >> SubCb() : %d %v", n, err)
}
