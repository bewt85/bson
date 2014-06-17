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
}{{
	v:        M{"int": int32(1)},
	expected: []byte("\b\x00\x00\x00int\x00\x01\x00\x00\x00\x00"),
}, {
	v:        M{"int": int64(1)},
	expected: []byte("\f\x00\x00\x00int\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00"),
}}

func TestMarshal(t *testing.T) {
	for _, tt := range marshalTests {
		got, err := Marshal(tt.v)
		if err != tt.err {
			t.Errorf("Marshal(%q): expected err: %v, got %v", tt.v, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("Marshal(%q): expected: %v, got: %v", tt.v, tt.expected, got)
		}
	}
}

var unmarshalTests = []struct {
	data     []byte
	expected interface{}
	err      error
}{{
	data: []byte{},
	err:  ErrTooShort,
}, {
	data: []byte{0x01},
	err:  ErrTooShort,
}, {
	data: []byte{0x05, 0x0, 0x0, 0x0},
	err:  ErrTooShort,
}, {
	data: []byte{0x04, 0x0, 0x0, 0x0},
	err:  ErrTooShort,
}, {
	data: []byte{0x04, 0x0, 0x0, 0x0, 0x0},
	err:  ErrTooShort,
}, {
	data:     []byte("\f\x00\x00\x00int\x00\x01\x00\x00\x00\x00"),
	expected: M{"int": int32(1)},
}, {
	data:     []byte("\x11\x00\x00\x00int\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00"),
	expected: M{"int": int64(1)},
}, {
	data:     []byte("\x16\x00\x00\x00\x02hello\x00\x06\x00\x00\x00world\x00\x00"),
	expected: M{"hello": "world"},
}}

func TestUnmarshal(t *testing.T) {
	for _, tt := range unmarshalTests {
		v := make(map[string]interface{})
		err := Unmarshal(tt.data, &v)
		if err != tt.err {
			t.Errorf("Unmarshal(%q): expected err: %v, got %v", tt.data, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, v) {
			t.Errorf("Unmarshal(%q): expected %q, got %q", tt.data, tt.expected, v)
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

func TestEncoderEncode(t *testing.T) {
	for _, tt := range marshalTests {
		var w bytes.Buffer
		e := NewEncoder(&w)
		err := e.Encode(tt.v)
		if err != tt.err {
			t.Errorf("Encoder.Encode(%q): expected err: %v, got %v", tt.v, tt.err, err)
			continue
		}
		got := w.Bytes()
		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("Encoder.Encode(%q): expected: %v, got: %v", tt.v, tt.expected, got)
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
	data:     []byte("\x0c\x00\x00\x00int\x00\x01\x00\x00\x00\x00"),
	expected: M{"int": int32(1)},
}, {
	data:     []byte("\x11\x00\x00\x00int\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00"),
	expected: M{"int": int64(1)},
}, {
	data:     []byte("\x16\x00\x00\x00\x02hello\x00\x06\x00\x00\x00world\x00\x00"),
	expected: M{"hello": "world"},
}}

func TestDecoderDecode(t *testing.T) {
	for _, tt := range decoderDecodeTests {
		r := bytes.NewReader(tt.data)
		d := NewDecoder(r)
		v := make(map[string]interface{})
		err := d.Decode(&v)
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("Decoder.Decode(%q): expected err: %v, got %v", tt.data, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, v) {
			t.Errorf("Decoder.Decode(%q): expected %q, got %q", tt.data, tt.expected, v)
		}
	}
}
