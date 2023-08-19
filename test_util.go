package main

import (
	"reflect"
	"runtime/debug"
	"testing"
)

func checkEq[V any](t *testing.T, a V, b V) {
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("got:%v, want:%v \n%s", a, b, debug.Stack())
	}
}
func checkNotEq[V any](t *testing.T, a V, b V) {
	if reflect.DeepEqual(a, b) {
		t.Fatalf("got:%v, want:%v \n%s", a, b, debug.Stack())
	}
}
