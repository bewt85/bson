package bson

import (
	"reflect"
	"testing"
)

type M map[string]interface{}

var encodeTests = []struct {
	m        map[string]interface{}
	expected []byte
}{{
	m:        M{"int": int32(1)},
	expected: []byte("\b\x00\x00\x00int\x00\x01\x00\x00\x00\x00"),
}, {
	m:        M{"int": int64(1)},
	expected: []byte("\f\x00\x00\x00int\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00"),
}}

func TestWriterWriteMap(t *testing.T) {
	for _, tt := range encodeTests {
		var w writer
		rv := reflect.ValueOf(tt.m)
		err := w.writeMap(rv)
		if err != nil {
			t.Errorf("writeMap(%q): %v", tt.m, err)
			continue
		}
		got := w.bson
		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("writeMap(%q): expected %q, got %q", tt.m, tt.expected, got)
			continue
		}
		// round trip
		r := reader{bson: got[4 : len(got)-1]}
		for r.Next() {
		}
		if err != nil {
			t.Errorf("writeMap(%q): round trip %v", tt.m, err)
		}
	}
}

func TestWriterGrow(t *testing.T) {
	checklen := func(b []byte, expected int) {
		if got := len(b); expected != got {
			t.Fatalf("checklen: expected %d, got %d", expected, got)
		}
	}
	checkcap := func(b []byte, expected int) {
		if got := cap(b); expected != got {
			t.Fatalf("checkcap: expected %d, got %d", expected, got)
		}
	}

	var w writer
	checklen(w.bson, 0)
	checkcap(w.bson, 0)

	var w2 = writer{bson: make([]byte, 5)}
	checklen(w2.bson, 5)
	checkcap(w2.bson, 5)
}
