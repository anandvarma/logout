package main

import (
	"math/rand"
	"time"
)

// logKey defines the key used to identify a log.
type logKey struct {
	owner int64
	token int64
}

// logVal defines the value held in the in-memory cache.
type logVal struct {
	endTimestamp int64
	rb           rollingBuf
	notif        string
}

// logState captures the in memory state of an actively streaming log.
type logState struct {
	logKey
	logVal
}

// NewLogState returns a new logState instance.
func newLogState() *logState {
	ls := logState{}
	ls.owner = ANONYMOUS_OWNER_ID
	ls.token = rand.Int63n(MAX_TOKEN_NUM)
	ls.endTimestamp = -1
	ls.rb = NewRollingBuf(BUFF_ARR_CAP)
	return &ls
}

// Implements the PubSub Subscriber interface.
func (ls *logState) SubCb(buf []byte) bool {
	if buf == nil {
		ls.endTimestamp = time.Now().Unix()
		return false
	}
	ls.rb.AddBuff(buf)
	return true // Remain subscribed.
}
