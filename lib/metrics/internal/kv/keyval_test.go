package kv

import (
	"reflect"
	"testing"
)

func TestKeyVal(t *testing.T) {
	var kv KeyVal

	kv.Append("alice", "1")
	kv.Append("eve", "3")
	kv.Append("bob", "2")

	kv.Sort()

	var (
		keysExpected = []string{"alice", "bob", "eve"}
		ValsExpected = []string{"1", "2", "3"}
	)

	if !reflect.DeepEqual(keysExpected, kv.Keys) {
		t.Errorf("keys not sorted, expected %#v, got %#v", keysExpected, kv.Keys)
	}

	if !reflect.DeepEqual(ValsExpected, kv.Vals) {
		t.Errorf("Vals not sorted, expected %#v, got %#v", ValsExpected, kv.Vals)
	}
}
