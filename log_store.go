package main

import (
	"errors"
	"log"
	"sync"
)

var ErrKeyExists = errors.New("insert will result in an overwrite")

type LogStore struct {
	cache map[int64]*RingBuf
	db    LogDb
	lock  sync.RWMutex
}

func NewLogStore() *LogStore {
	var dbImpl LogDb = nil
	if len(LOG_DB_PATH) == 0 {
		log.Println("Using Noop LogDb")
		dbImpl = createNoopLogDb()
	} else {
		log.Printf("Using FS LogDb at: %s", LOG_DB_PATH)
		dbImpl = createFSLogDb(LOG_DB_PATH)
	}
	return &LogStore{
		cache: make(map[int64]*RingBuf),
		db:    dbImpl,
		lock:  sync.RWMutex{},
	}
}

func (lc *LogStore) MemPut(key int64, rb *RingBuf) error {
	lc.lock.Lock()
	defer lc.lock.Unlock()

	_, exists := lc.cache[key]
	if exists {
		return ErrKeyExists
	}

	lc.cache[key] = rb
	return nil
}

func (lc *LogStore) MemGet(key int64) (*RingBuf, bool) {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	val, exists := lc.cache[key]
	return val, exists
}

func (lc *LogStore) Persist(key int64) {
	rb, exists := lc.MemGet(key)
	if !exists {
		log.Printf("%d doesn't exist, cannot persist!", key)
		return
	}

	// Depending on the implementation of Commit, it may or may not block.
	// To avoid starving other potentially much faster operations, we do not
	// grab a lock for this section.
	ok := lc.db.Commit(key, rb)
	if !ok {
		return
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()
	delete(lc.cache, key)
}
