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
	bson: []byte("\x16\x00\x00\x00\x02hello\x00\x06\x00\x00\x00world\x00\x00"),
	expected: []element{{
		typ:     0x02, // utf-8 string
		ename:   cstring("hello"),
		element: []byte("world\x00"),
	}},
	err: nil,
}, {
	bson: []byte("\x31\x00\x00\x00\x04BSON\x00\x26\x00\x00\x00\x020\x00\x08\x00\x00\x00awesome\x00\x011\x00\x33\x33\x33\x33\x33\x33\x14\x40\x102\x00\xc2\x07\x00\x00\x00\x00"),
	expected: []element{{
		typ:     0x4, // bson array
		ename:   cstring("BSON"),
		element: []byte("\x26\x00\x00\x00\x020\x00\x08\x00\x00\x00awesome\x00\x011\x00\x33\x33\x33\x33\x33\x33\x14\x40\x102\x00\xc2\x07\x00\x00\x00"),
	}},
	err: nil,
}, {
	// test1.bson
	bson: []byte{0x0e, 0x00, 0x00, 0x00, 0x10, 0x69, 0x6e, 0x74, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
	expected: []element{{
		typ:     0x10,
		ename:   cstring("int"),
		element: []byte{0x01, 0, 0, 0},
	}},
}, {
	// test2.bson
	bson: []byte{0x14, 0x00, 0x00, 0x00, 0x12, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	expected: []element{{
		typ:     0x12,
		ename:   cstring("int64"),
		element: []byte{0x01, 0, 0, 0, 0, 0, 0, 0},
	}},
}, {
	// test3.bson
	bson: []byte{0x15, 0x00, 0x00, 0x00, 0x01, 0x64, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x00, 0x2b, 0x87, 0x16, 0xd9, 0xce, 0xf7, 0xf1, 0x3f, 0x00},
	expected: []element{{
		typ:     0x1,
		ename:   cstring("double"),
		element: []uint8{0x2b, 0x87, 0x16, 0xd9, 0xce, 0xf7, 0xf1, 0x3f},
	}},
}}

func TestBsonIter(t *testing.T) {
	for _, tt := range bsonIterTests {
		iter := bsonIter{bson: tt.bson[4 : len(tt.bson)-1]}
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
