package bson

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
)

var marshalTests = []struct {
	v        interface{}
	expected []byte
	err      error
}{}

func TestMarshal(t *testing.T) {
	for _, tt := range marshalTests {
		got, err := Marshal(tt.v)
		if err != tt.err {
			t.Errorf("Marshal(%#v): expected err: %v, got %v", tt.v, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("Marshal(%#v): expected: %v, got: %v", tt.v, tt.expected, got)
		}
	}
}

var unmarshalTests = []struct {
	data     []byte
	expected interface{}
	err      error
}{}

func TestUnmarshal(t *testing.T) {
	for _, tt := range unmarshalTests {
		var v interface{}
		err := Unmarshal(tt.data, &v)
		if err != tt.err {
			t.Errorf("Unmarshal(%v): expected err: %v, got %v", tt.data, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, v) {
			t.Errorf("Unmarshal(%v): expected %v, got %v", tt.data, tt.expected, v)
		}
	}
}

func TestNewDecoder(t *testing.T) {
	var r bytes.Buffer
	var d interface{}
	d = NewDecoder(&r)
	switch d.(type) {
	case *Decoder:
	default:
		t.Fatal("NewDecoder extended %T, got %T", new(Decoder), d)
	}
	if d == nil {
		t.Fatal("NewDecoder returned nil *Decoder")
	}
}

func TestNewEncoder(t *testing.T) {
	var w bytes.Buffer
	var e interface{}
	e = NewEncoder(&w)
	switch e.(type) {
	case *Encoder:
	default:
		t.Fatal("NewEncoder extended %T, got %T", new(Encoder), e)
	}
	if e == nil {
		t.Fatal("NewEncoder returned nil *Encoder")
	}
}

var encoderEncodeTests = []struct {
	v        interface{}
	expected []byte
	err      error
}{}

func TestEncoderEncode(t *testing.T) {
	for _, tt := range encoderEncodeTests {
		var w bytes.Buffer
		e := NewEncoder(&w)
		err := e.Encode(tt.v)
		if err != tt.err {
			t.Errorf("Encoder.Encode(%#v): expected err: %v, got %v", tt.v, tt.err, err)
			continue
		}
		got := w.Bytes()
		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("Encoder.Encode(%#v): expected: %v, got: %v", tt.v, tt.expected, got)
		}
	}
}

var decoderDecodeTests = []struct {
	data     []byte
	expected interface{}
	err      error
}{{
	data: []byte{},
	err:  io.EOF,
}, {
	data: []byte{0x01},
	err:  io.ErrUnexpectedEOF,
}, {
	data: []byte{0x05, 0x0, 0x0, 0x0},
	err:  io.ErrUnexpectedEOF,
}, {
	data: []byte{0x04, 0x0, 0x0, 0x0},
	err:  ErrTooShort,
}, {
	data: []byte{0x04, 0x0, 0x0, 0x0, 0x0},
	err:  ErrTooShort,
}}

func TestDecoderDecode(t *testing.T) {
	for _, tt := range decoderDecodeTests {
		r := bytes.NewReader(tt.data)
		d := NewDecoder(r)
		var v interface{}
		err := d.Decode(&v)
		if err != tt.err {
			t.Errorf("Decoder.Decode(%v): expected err: %v, got %v", tt.data, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, v) {
			t.Errorf("Decoder.Decode(%v): expected %v, got %v", tt.data, tt.expected, v)
		}
	}
}

var readInt32Tests = []struct {
	data     []byte
	expected int
	rest     []byte
}{{
	data:     []byte{0x1, 0x1, 0x0, 0x0},
	expected: 0x101,
	rest:     []byte{},
}, {
	data:     []byte{0x0, 0x0, 0x0, 0x1},
	expected: 0x01000000,
	rest:     []byte{},
}, {
	data:     []byte{0x0f, 0x0f, 0x0f, 0x0f, 0x0f},
	expected: 0x0f0f0f0f,
	rest:     []byte{0x0f},
}}

func TestReadInt32(t *testing.T) {
	for _, tt := range readInt32Tests {
		got, rest := readInt32(tt.data)
		if got != tt.expected || !reflect.DeepEqual(tt.rest, rest) {
			t.Errorf("readInt32(%v): expected %v %v, got %v, %v", tt.data, tt.expected, tt.rest, got, rest)
		}
	}
}

func cstring(s string) []byte {
	return append([]byte(s), 0)
}

var readCstringTests = []struct {
	data           []byte
	expected, rest []byte
	err            error
}{{
	data:     []byte{},
	expected: nil,
	rest:     nil,
	err:      errors.New("bson: cstring missing \\0"),
}, {
	data:     cstring("bson"),
	expected: cstring("bson"),
	rest:     []byte{},
	err:      nil,
}, {
	data:     cstring("bson\x00"),
	expected: cstring("bson"),
	rest:     []byte{0},
	err:      nil,
}}

func TestReadCstring(t *testing.T) {
	for _, tt := range readCstringTests {
		got, rest, err := readCstring(tt.data)
		if !reflect.DeepEqual(tt.err, err) || !reflect.DeepEqual(tt.expected, got) || !reflect.DeepEqual(tt.rest, rest) {
			t.Errorf("readCstring(%v): expected %v %v %v, got %v %v %v", tt.data, tt.expected, tt.rest, tt.err, got, rest, err)
		}
	}
}
