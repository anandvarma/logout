package main

import "io"

type StreamLogOp struct {
	out     io.Writer
	logBufs [][]byte
}

func NewStreamLogOp(out io.Writer) *StreamLogOp {
	return &StreamLogOp{
		out:     out,
		logBufs: make([][]byte, 0 /* size */, BUFF_ARR_CAP /* capacity */),
	}
}

func (op *StreamLogOp) Subscribe(ba *BuffArray) {

}
