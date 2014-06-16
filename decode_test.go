package bson

import (
	"reflect"
	"testing"
)

var decodeTests = []struct {
	data []byte
	v    interface{}
	err  error
}{{
	data: []byte{0x05, 0x0, 0x0, 0x0},
	err:  ErrTooShort,
}, {
	data: []byte{0x05, 0x0, 0x0, 0x0, 0x0},
	v:    struct{}{},
	err:  &InvalidUnmarshalError{reflect.TypeOf(struct{}{})},
}}

func TestDecode(t *testing.T) {
	for _, tt := range decodeTests {
		v := tt.v
		err := decode(tt.data, v)
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("decode(%v): expected err %v, got %v", tt.data, tt.err, err)
			continue
		}
	}
}

type element struct {
	typ     byte
	ename   []byte
	element []byte
}

var bsonIterTests = []struct {
	bson     []byte
	expected []element
	err      error
}{{
	bson:     []byte{},
	expected: []element{},
	err:      nil,
}, {
	bson: []byte("\x02hello\x00\x06\x00\x00\x00world\x00"),
	expected: []element{{
		typ:     0x02, // utf-8 string
		ename:   cstring("hello"),
		element: []byte("world\x00"),
	}},
	err: nil,
}}

func TestBsonIter(t *testing.T) {
	for _, tt := range bsonIterTests {
		iter := bsonIter{bson: tt.bson}
		got := make([]element, 0)
		for iter.Next() {
			typ, ename, value := iter.Element()
			got = append(got, element{typ, ename, value})
		}
		err := iter.Err()
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("bsonIter %v: expected err %v, got %v", tt.bson, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("bsonIter %v: expected %#v, got %#v", tt.bson, tt.expected, got)
		}
	}
}
