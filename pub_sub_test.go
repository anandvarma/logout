package main

import (
	"math/rand"
	"testing"
)

var totalReads int = 0

type DummySub struct {
}

func (sub *DummySub) SubCb(buf []byte) bool {
	totalReads++
	return true
}

func TestPubSub(t *testing.T) {
	ps := NewPubSub()
	totalReads = 0

	// Initial state sanity test.
	if len(ps.bus) != 0 {
		t.Errorf("Unexpected pub sub bus size:%d", len(ps.bus))
	}

	// Publishes without any subs should end up being no ops.
	ps.Publish(1, []byte("Foo"))
	if totalReads != 0 {
		t.Errorf("Spurious publish")
	}

	// Unsubscribe of a non existing topic.
	err := ps.Unsubscribe(1, &DummySub{})
	if err != ErrTopicNotFound {
		t.Errorf("Got back unexpected error: %v", err)
	}

	// Add a sub to the topic '1'.
	ps.Subscribe(1, &DummySub{})
	if len(ps.bus) != 1 {
		t.Errorf("Unexpected bus size: %d", len(ps.bus))
	}

	// Publish and check subscription.
	for i := 0; i < 16; i++ {
		ps.Publish(1, []byte("Bar"))
		ps.Publish(100, []byte("Bar")) // noop
	}
	if totalReads != 16 {
		t.Errorf("Unexpected number of reads: %d", totalReads)
	}

	// Unsubscribe a non existent sub.
	err = ps.Unsubscribe(1, &DummySub{})
	if err != ErrSubNotFound {
		t.Errorf("Got back unexpected error: %v", err)
	}

	// Unsusbscribe existing sub.
	err = ps.Unsubscribe(1, ps.bus[1][0])
	if err != nil {
		t.Errorf("Got back unexpected error: %v", err)
	}
	if len(ps.bus) != 0 {
		t.Errorf("Unexpected pub sub bus size:%d", len(ps.bus))
	}

	// Subsequent publishes should be noops.
	ps.Publish(1, []byte("Bar"))
	ps.Publish(100, []byte("Bar")) // noop
	if totalReads != 16 {
		t.Errorf("Unexpected number of reads: %d", totalReads)
	}

	// Test multiple topics and subs.
	for i := int64(0); i < 64; i++ {
		for j := 0; j < 8; j++ {
			ps.Subscribe(i, &DummySub{})
		}
	}
	if len(ps.bus) != 64 {
		t.Errorf("Unexpected bus size: %d", len(ps.bus))
	}
	for i := int64(0); i < 64; i++ {
		if len(ps.bus[i]) != 8 {
			t.Errorf("Unexpected sub list size at: %d", i)
		}
	}

	totalReads = 0
	for i := int64(0); i < 64; i++ {
		ps.Publish(i, []byte("Bar"))
	}
	if totalReads != 64*8 {
		t.Errorf("Unexpected number of reads: %d", totalReads)
	}

	// Delete subs.
	for i := int64(0); i < 64; i++ {
		for j := 0; j < 8; j++ {
			err = ps.Unsubscribe(i, ps.bus[i][0])
			if err != nil {
				t.Errorf("Got back unexpected error: %v", err)
			}
		}
	}

	if len(ps.bus) != 0 {
		t.Errorf("Unexpected pub sub bus size:%d", len(ps.bus))
	}
}

type DummySubOneShot struct {
}

func (sub *DummySubOneShot) SubCb(buf []byte) bool {
	totalReads++
	return false
}

func TestPubSubOneShot(t *testing.T) {
	ps := NewPubSub()
	totalReads = 0

	// Add a sub to the topic '1'.
	ps.Subscribe(1, &DummySubOneShot{})
	if len(ps.bus) != 1 {
		t.Errorf("Unexpected bus size: %d", len(ps.bus))
	}

	// Publish and check subscription. Only one event should have fired.
	for i := 0; i < 16; i++ {
		ps.Publish(1, []byte("Bar"))
		ps.Publish(100, []byte("Bar")) // noop
	}
	if totalReads != 1 {
		t.Errorf("Unexpected number of reads: %d", totalReads)
	}

	// Oneshot should have gotten unsubscribed already.
	if len(ps.bus) != 0 {
		t.Errorf("Unexpected bus size: %d", len(ps.bus))
	}

	// Test multiple topics and subs.
	for i := int64(0); i < 64; i++ {
		for j := 0; j < 8; j++ {
			ps.Subscribe(i, &DummySubOneShot{})
		}
	}
	if len(ps.bus) != 64 {
		t.Errorf("Unexpected bus size: %d", len(ps.bus))
	}
	for i := int64(0); i < 64; i++ {
		if len(ps.bus[i]) != 8 {
			t.Errorf("Unexpected sub list size at: %d", i)
		}
	}

	totalReads = 0
	randOrder := rand.Perm(64)
	t.Log(randOrder)
	for _, i := range randOrder {
		ps.Publish(int64(i), []byte("Bar"))
	}
	if totalReads != 64*8 {
		t.Errorf("Unexpected number of reads: %d", totalReads)
	}

	if len(ps.bus) != 0 {
		t.Errorf("Unexpected pub sub bus size:%d", len(ps.bus))
	}
}
