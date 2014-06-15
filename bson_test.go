package bson

import (
	"bytes"
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
}, {
	data: []byte{0x05, 0x0, 0x0, 0x0, 0x0},
	err:  nil,
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
