package main

import (
	"io"
	"log"
)

type RingBuf struct {
	arr        [][]byte
	numAppends int
	isClosed   bool
}

func NewRingBuf(capacity int) *RingBuf {
	return &RingBuf{
		arr:        make([][]byte, 0 /* size */, capacity),
		numAppends: 0,
		isClosed:   false,
	}
}

func (ba *RingBuf) Len() int {
	if ba.numAppends < cap(ba.arr) {
		return ba.numAppends
	} else {
		return cap(ba.arr)
	}
}

// PubSub interface.
func (ba *RingBuf) SubCb(buf []byte) {
	log.Printf("RingBuf sub << %d", len(buf))
	ba.AddBuff(buf)
}

func (ba *RingBuf) AddBuff(buff []byte) {
	// Guardrail to prevent writes after Close.
	if ba.isClosed {
		log.Fatal("Writing to RingBuf after Close")
	}

	capOld := cap(ba.arr)
	overWrite := ba.numAppends >= cap(ba.arr)
	if !overWrite {
		ba.arr = append(ba.arr, buff)
	} else {
		writeIdx := ba.numAppends % cap(ba.arr)
		ba.arr[writeIdx] = buff
	}
	ba.numAppends++

	// Assertion.
	if cap(ba.arr) != capOld {
		log.Fatal("RingBuf grew in size")
	}
}

func (ba *RingBuf) Close() {
	log.Println("RingBuf Close()")
	ba.isClosed = true
}

func (ba *RingBuf) IsClosed() bool {
	return ba.isClosed
}

func (ba *RingBuf) Read(buff []byte) (n int, err error) {
	// Guardrail to ensure Reads come in only after Close.
	if !ba.isClosed {
		log.Fatal("Reading from RingBuf before Close")
	}

	readBytes := 0
	startIdx := ba.numAppends % cap(ba.arr)
	for ii := startIdx; ii < ba.Len(); ii++ {
		readBytes += copy(buff[readBytes:], ba.arr[ii])
	}
	for ii := 0; ii < startIdx; ii++ {
		readBytes += copy(buff[readBytes:], ba.arr[ii])
	}
	return readBytes, io.EOF
}
