package main

import (
	"testing"
)

func TestBoltConfStore(t *testing.T) {
	cs := createBoltConfDB("./db-test")
	checkNotEq(t, cs.db, nil)

	testConfStore(t, cs)
}

func testConfStore(t *testing.T, cs ConfStore) {
}
