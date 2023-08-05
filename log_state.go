package main

import (
	"log"
	"math/rand"
	"time"
)

// logState captures the in memory state of an actively streaming log.
type logState struct {
	token        int64
	endTimestamp int64
	rb           rollingBuf
	// lock?
}

// NewLogState returns a new logState instance.
func newLogState() *logState {
	return &logState{
		token:        rand.Int63n(MAX_TOKEN_NUM),
		endTimestamp: -1,
		rb:           NewRollingBuf(BUFF_ARR_CAP),
	}
}

// Implements the PubSub Subscriber interface.
func (ls *logState) SubCb(buf []byte) bool {
	if buf == nil {
		ls.endTimestamp = time.Now().Unix()
		log.Printf("EOF: %v", ls)
		return false
	}
	ls.rb.AddBuff(buf)
	return true // Remain subscribed.
}

// Returns true if the log is actively being drained.
func (ls *logState) IsActive() bool {
	return ls.endTimestamp < 0
}
