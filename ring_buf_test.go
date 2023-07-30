package main

import (
	"testing"
)

func TestRingBuf(t *testing.T) {
	rb := NewRingBuf(2)

	// Sanity check.
	checkEq(t, cap(rb.arr), 2)
	checkEq(t, totalLen(rb), 0)
	checkEq(t, rb.IsClosed(), false)
	checkEq(t, rb.Len(), 0)

	// Fill the ring buffer.
	rb.AddBuff([]byte{1, 0, 0})
	rb.AddBuff([]byte{2, 0, 0})

	checkEq(t, cap(rb.arr), 2)
	checkEq(t, totalLen(rb), 6)
	checkEq(t, rb.IsClosed(), false)
	checkEq(t, rb.Len(), 2)
	readBuf := make([]byte, 6)
	rb.Read(readBuf)
	checkEq(t, readBuf, []byte{1, 0, 0, 2, 0, 0})

	// Overflow the ring buffer.
	rb.AddBuff([]byte{3, 0, 0})

	checkEq(t, cap(rb.arr), 2)
	checkEq(t, totalLen(rb), 6)
	checkEq(t, rb.IsClosed(), false)
	checkEq(t, rb.Len(), 2)
	rb.Read(readBuf)
	checkEq(t, readBuf, []byte{2, 0, 0, 3, 0, 0})

	// Close ring buffer.
	rb.AddBuff(nil)
	checkEq(t, rb.IsClosed(), true)
}

func totalLen(rb *RingBuf) int {
	ret := 0
	for _, i := range rb.arr {
		ret += len(i)
	}
	return ret
}
