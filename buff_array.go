package main

import (
	"io"
	"log"
)

type BuffArray struct {
	arr        [][]byte
	cap        int
	numAppends int
}

func NewBuffArray(capacity int) *BuffArray {
	return &BuffArray{
		arr:        make([][]byte, 0 /* size */, capacity),
		cap:        capacity,
		numAppends: 0,
	}
}

func (ba *BuffArray) Len() int {
	if ba.numAppends < ba.cap {
		return ba.numAppends
	} else {
		return ba.cap
	}
}

func (ba *BuffArray) AddBuff(buff []byte) {
	overWrite := ba.numAppends >= ba.cap
	if !overWrite {
		ba.arr = append(ba.arr, buff)
	} else {
		writeIdx := ba.numAppends % ba.cap
		ba.arr[writeIdx] = buff
	}
	ba.numAppends++

	// Assertion.
	if cap(ba.arr) != ba.cap {
		log.Fatal("BuffArray grew in size")
	}
}

func (ba *BuffArray) Read(buff []byte) (n int, err error) {
	readBytes := 0
	startIdx := ba.numAppends % ba.cap
	for ii := startIdx; ii < ba.Len(); ii++ {
		readBytes += copy(buff[readBytes:], ba.arr[ii])
	}
	for ii := 0; ii < startIdx; ii++ {
		readBytes += copy(buff[readBytes:], ba.arr[ii])
	}
	return readBytes, io.EOF
}
