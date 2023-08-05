package main

import (
	"errors"
	"log"
	"sync"
)

var ErrKeyExists = errors.New("insert will result in an overwrite")

type LogStore struct {
	cache map[int64]*logState
	db    LogDb
	lock  sync.RWMutex // Can remove? All indexes are random unless we hit collisions.
}

func NewLogStore() LogStore {
	var dbImpl LogDb = nil
	if len(LOG_DB_PATH) == 0 {
		log.Println("Using Noop LogDb")
		dbImpl = createNoopLogDb()
	} else {
		log.Printf("Using FS LogDb at: %s", LOG_DB_PATH)
		dbImpl = createFSLogDb(LOG_DB_PATH)
	}
	return LogStore{
		cache: make(map[int64]*logState),
		db:    dbImpl,
		lock:  sync.RWMutex{},
	}
}

func (store *LogStore) MemPut(ls *logState) error {
	store.lock.Lock()
	defer store.lock.Unlock()

	_, exists := store.cache[ls.token]
	if exists {
		return ErrKeyExists
	}

	store.cache[ls.token] = ls
	return nil
}

func (lc *LogStore) MemGet(key int64) (*logState, bool) {
	lc.lock.RLock()
	defer lc.lock.RUnlock()

	val, exists := lc.cache[key]
	return val, exists
}

func (lc *LogStore) Persist(key int64) {
	ls, exists := lc.MemGet(key)
	if !exists {
		log.Printf("%d doesn't exist, cannot persist!", key)
		return
	}

	// Depending on the implementation of Commit, it may or may not block.
	// To avoid starving other potentially much faster operations, we do not
	// grab a lock for this section.
	ok := lc.db.Commit(*ls)
	if !ok {
		return
	}

	lc.lock.Lock()
	defer lc.lock.Unlock()
	delete(lc.cache, key)
}
