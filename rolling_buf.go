package main

import (
	"io"
)

type rollingBuf struct {
	arr        [][]byte
	numAppends int
}

func NewRollingBuf(capacity int) rollingBuf {
	return rollingBuf{
		arr:        make([][]byte, 0 /* size */, capacity),
		numAppends: 0,
	}
}

func (ba *rollingBuf) Len() int {
	if ba.numAppends < cap(ba.arr) {
		return ba.numAppends
	} else {
		return cap(ba.arr)
	}
}

func (ba *rollingBuf) AddBuff(buff []byte) {
	overWrite := ba.numAppends >= cap(ba.arr)
	if !overWrite {
		ba.arr = append(ba.arr, buff)
	} else {
		writeIdx := ba.numAppends % cap(ba.arr)
		ba.arr[writeIdx] = buff
	}
	ba.numAppends++
}

func (ba *rollingBuf) ForEachBuf(fn func([]byte)) {
	startIdx := ba.numAppends % cap(ba.arr)
	for ii := startIdx; ii < ba.Len(); ii++ {
		fn(ba.arr[ii])
	}
	for ii := 0; ii < startIdx; ii++ {
		fn(ba.arr[ii])
	}
}

// Reader interface.
func (ba *rollingBuf) Read(buff []byte) (n int, err error) {
	readBytes := 0
	ba.ForEachBuf(func(b []byte) {
		readBytes += copy(buff[readBytes:], b)
	})
	return readBytes, io.EOF
}

// Closer interface.
func (ba *rollingBuf) Close() error {
	return nil
}
