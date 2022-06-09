package main

import (
	"io"
	"log"
	"sync"
)

type RingBuf struct {
	arr        [][]byte
	numAppends int
	isClosed   bool
	lock       sync.RWMutex
}

func NewRingBuf(capacity int) *RingBuf {
	return &RingBuf{
		arr:        make([][]byte, 0 /* size */, capacity),
		numAppends: 0,
		isClosed:   false,
		lock:       sync.RWMutex{},
	}
}

func (ba *RingBuf) Len() int {
	ba.lock.RLock()
	defer ba.lock.RUnlock()

	if ba.numAppends < cap(ba.arr) {
		return ba.numAppends
	} else {
		return cap(ba.arr)
	}
}

// PubSub interface.
func (ba *RingBuf) SubCb(buf []byte) bool {
	log.Printf("RingBuf sub << %d", len(buf))
	ba.AddBuff(buf)
	return true // Remain subscribed.
}

func (ba *RingBuf) AddBuff(buff []byte) {
	ba.lock.Lock()
	defer ba.lock.Unlock()

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
	ba.lock.Lock()
	defer ba.lock.Unlock()

	log.Println("RingBuf Close()")
	ba.isClosed = true
}

func (ba *RingBuf) IsClosed() bool {
	ba.lock.RLock()
	defer ba.lock.RUnlock()

	return ba.isClosed
}

func (ba *RingBuf) ForEachBuf(fn func([]byte)) {
	ba.lock.RLock()
	defer ba.lock.RUnlock()

	// Guardrail to ensure Reads come in only after Close.
	if !ba.isClosed {
		//log.Fatal("Reading from RingBuf before Close")
	}

	startIdx := ba.numAppends % cap(ba.arr)
	for ii := startIdx; ii < ba.Len(); ii++ {
		fn(ba.arr[ii])
	}
	for ii := 0; ii < startIdx; ii++ {
		fn(ba.arr[ii])
	}
}

// Reader interface.
func (ba *RingBuf) Read(buff []byte) (n int, err error) {
	readBytes := 0
	ba.ForEachBuf(func(b []byte) {
		readBytes += copy(buff[readBytes:], b)
	})
	return readBytes, io.EOF
}
