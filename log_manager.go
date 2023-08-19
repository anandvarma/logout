package main

import (
	"io"
	"sync"
)

type logManager struct {
	cache map[logKey]*logVal
	lock  sync.RWMutex
	ls    LogStore
	cs    ConfStore
}

func newLogManager(ls LogStore, cs ConfStore) logManager {
	return logManager{
		cache: make(map[logKey]*logVal),
		lock:  sync.RWMutex{},
		ls:    ls,
		cs:    cs,
	}
}

// OpenLog tracks an open log with 'key' and state defined in 'val'.
func (lm *logManager) OpenLog(key logKey, val *logVal) error {
	lm.lock.Lock()
	defer lm.lock.Unlock()

	_, exists := lm.cache[key]
	if exists {
		return ErrKeyExists
	}

	lm.cache[key] = val
	return nil
}

// GetOpenLogReader returns an io.Reader for an open log with 'key'.
func (lm *logManager) GetOpenLogReader(key logKey) (io.Reader, error) {
	val, err := lm.getLogVal(key)
	if err != nil {
		return nil, err
	}
	return &val.rb, nil
}

// CloseLog transitions 'key' from an open to a closed state.
// An open log is an actively draining that is cached in memory.
// A closed log is one that is done draining and is immutable.
// This log is persisted to a logStore and removed from the cache.
func (lm *logManager) CloseLog(key logKey) error {
	val, err := lm.getLogVal(key)
	if err != nil {
		return err
	}

	// Depending on the implementation of Commit, it may or may not block.
	// To avoid starving other potentially much faster operations, we do not
	// grab a lock for this section.
	err = lm.ls.CommitLog(key, *val)
	if err != nil {
		return err
	}

	lm.lock.RLock()
	defer lm.lock.RUnlock()
	delete(lm.cache, key)
	return nil
}

func (lm *logManager) SetLabel(key logKey, label string) error {
	lm.lock.Lock()
	defer lm.lock.Unlock()

	val, exists := lm.cache[key]
	if !exists {
		return ErrInvalidKey
	}

	val.label = label
	return nil
}

func (lm *logManager) GetLabel(key logKey) (string, error) {
	val, err := lm.getLogVal(key)
	if err != nil {
		return "", err
	}
	return val.label, nil
}

func (lm *logManager) getLogVal(key logKey) (*logVal, error) {
	lm.lock.RLock()
	defer lm.lock.RUnlock()

	val, exists := lm.cache[key]
	if exists {
		return val, nil
	} else {
		return nil, ErrLogNotFound
	}
}
