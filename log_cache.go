package main

import (
	"errors"
	"fmt"
	"sync"
)

var ErrKeyExists = errors.New("insert will result in an overwrite")

type LogCache struct {
	cache map[int64]*RingBuf
	lock  sync.RWMutex
}

func NewLogCache() *LogCache {
	return &LogCache{
		cache: make(map[int64]*RingBuf, CACHE_SIZE),
		lock:  sync.RWMutex{},
	}
}

func (lc *LogCache) Put(key int64, rb *RingBuf) error {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	_, exists := lc.cache[key]
	if exists {
		return ErrKeyExists
	}

	lc.cache[key] = rb
	return nil
}

func (lc *LogCache) Get(key int64) (*RingBuf, bool) {
	val, exists := lc.cache[key]
	return val, exists
}

func (lc *LogCache) DebugString() string {
	return fmt.Sprintf("%v", lc.cache)
}
