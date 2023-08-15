package main

import (
	"errors"
	"sync"
)

type Subscriber interface {
	// Callback invoked when a new event is to be dispatched.
	// Returns true if interested in more events. Returning false unsubscribes.
	// This call is expected to be non blocking and quick.
	// This call should not directly call any PubSub methods. Doing so may
	// result in dead locks.
	SubCb(buf []byte) bool
}

// TODO: Make generic for Go 1.18+

var ErrTopicNotFound = errors.New("no such topic in the pub-sub bus")
var ErrSubNotFound = errors.New("no such sub in the pub-sub bus")

type PubSub struct {
	bus  map[logKey][]Subscriber
	lock sync.RWMutex
}

func NewPubSub() PubSub {
	return PubSub{
		bus:  make(map[logKey][]Subscriber),
		lock: sync.RWMutex{},
	}
}

func (ps *PubSub) Publish(key logKey, val []byte) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	subs, exists := ps.bus[key]
	if !exists || len(subs) == 0 {
		return
	}

	for _, sub := range subs {
		cont := sub.SubCb(val)
		if !cont {
			// Subscriber wishes to unsubscribe from further events.
			// Defer to avoid mutating slice mid range iteration.
			defer ps.unsubscribeUnsafe(key, sub)
		}
	}
}

func (ps *PubSub) Subscribe(key logKey, sub Subscriber) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	subs, exists := ps.bus[key]
	if !exists {
		subs = make([]Subscriber, 0 /* size */)
		ps.bus[key] = subs
	}

	subs = append(subs, sub)
	ps.bus[key] = subs
}

func (ps *PubSub) Unsubscribe(key logKey, sub Subscriber) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	return ps.unsubscribeUnsafe(key, sub)
}

func (ps *PubSub) unsubscribeUnsafe(key logKey, sub Subscriber) error {
	subs, exists := ps.bus[key]
	if !exists || len(subs) == 0 {
		return ErrTopicNotFound
	}

	i, err := findSub(sub, subs)
	if err != nil {
		return err
	}

	if len(subs) == 1 {
		// Last sub.
		delete(ps.bus, key)
		return nil
	}
	// Optimized unstable delete.
	subs[i] = subs[len(subs)-1]
	subs = subs[:len(subs)-1]
	ps.bus[key] = subs
	return nil
}

func findSub(sub Subscriber, subs []Subscriber) (int, error) {
	for i, s := range subs {
		if s == sub {
			return i, nil
		}
	}
	return -1, ErrSubNotFound
}
