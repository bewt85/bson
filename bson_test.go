package bson

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

var marshalTests = []struct {
	v        interface{}
	expected []byte
	err      error
}{{
	v:        M{"int": int32(1)},
	expected: []byte("\x0e\x00\x00\x00\x10int\x00\x01\x00\x00\x00\x00"),
}, {
	v:        M{"int64": int64(1)},
	expected: []byte{0x14, 0x00, 0x00, 0x00, 0x12, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}}

func TestMarshal(t *testing.T) {
	for _, tt := range marshalTests {
		got, err := Marshal(tt.v)
		if err != tt.err {
			t.Errorf("Marshal(%#v): expected err: %v, got %v", tt.v, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, got) {
			t.Errorf("Marshal(%#v): expected: % #x, got: % #x", tt.v, tt.expected, got)
		}
	}
}

var unmarshalTests = []struct {
	data     []byte
	expected interface{}
	err      error
}{{
	data:     []byte("\x0e\x00\x00\x00\x10int\x00\x01\x00\x00\x00\x00"),
	expected: map[string]interface{}{"int": int32(1)},
}, {
	data:     []byte{0x14, 0x00, 0x00, 0x00, 0x12, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	expected: map[string]interface{}{"int64": int64(1)},
}, {
	data:     []byte("\x16\x00\x00\x00\x02hello\x00\x06\x00\x00\x00world\x00\x00"),
	expected: map[string]interface{}{"hello": "world"},
}}

func TestUnmarshal(t *testing.T) {
	for _, tt := range unmarshalTests {
		v := make(map[string]interface{})
		err := Unmarshal(tt.data, &v)
		if err != nil {
			t.Errorf("Unmarshal(% #x): expected err: %v, got %v", tt.data, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, v) {
			t.Errorf("Unmarshal(%v): expected %# x, got %# x", tt.data, tt.expected, v)
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
			t.Errorf("Encoder.Encode(%#v): expected: %# x, got: %# x", tt.v, tt.expected, got)
		}
	}
}

func TestDecoderDecode(t *testing.T) {
	for _, tt := range unmarshalTests {
		r := bytes.NewReader(tt.data)
		d := NewDecoder(r)
		v := make(map[string]interface{})
		err := d.Decode(&v)
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("Decoder.Decode(% #x): expected err: %v, got %v", tt.data, tt.err, err)
			continue
		}
		if !reflect.DeepEqual(tt.expected, v) {
			t.Errorf("Decoder.Decode(%q): expected %q, got %q", tt.data, tt.expected, v)
		}
	}
}

var libbsonTests = []string{
	"test1.bson",
	"test2.bson",
	// "test3.bson", // double
	//"test4.bson", // timestamp
	"test5.bson",
	"test6.bson",
	// "test7.bson", // []double
	"test8.bson",
	"test9.bson", // null
	// "test10.bson", // regex
	"test11.bson",
	//"test12.bson", // bson is awesome
	"test13.bson", // array[bool]
	"test14.bson", // array[string]
	// "test15.bson", // array[datetime]
	"test16.bson",
	// "test17.bson", // objectid
	"test18.bson", // map[nil]
	"test19.bson",
	// "test20.bson",
	"test21.bson",
	"test23.bson",
	// "test24.bson", // binary data
	//"test25.bson", "test32.bson" // deprecated
	//"test26.bson", // datatime
	// "test27.bson", // regex
	// "test28.bson", // db pointer
	// "test29.bson", "test30.bson", // javascript
	// "test31.bson", // javascript w/scope
	//"test33.bson",
	// "test34.bson", // one byte short ...
	// "test35.bson", // timestamp
	// "test36.bson", // MinKey
	// "test37.bson", // MaxKey
	"test38.bson",
	"test39.bson",
	"dollarquery.bson",
	// "dotquery.bson",
	// "dotkey.bson",
	// "stackoverflow.bson",
	"test56.bson",
	"readergrow.bson",
}

// round trip the data in testdata/ taken from the libbson tests.
func TestLibBSONTestdata(t *testing.T) {
	for _, tt := range libbsonTests {
		f := filepath.Join("testdata", tt)
		want, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatal(f, err)
		}
		d := NewDecoder(bytes.NewReader(want))
		v := make(map[string]interface{})
		if err := d.Decode(&v); err != nil {
			t.Errorf("Decode (%# 0x): %s: %v", want, f, err)
			continue
		}
		var out bytes.Buffer
		e := NewEncoder(&out)
		if err := e.Encode(v); err != nil {
			t.Errorf("Encode: %s: %v", f, err)
			continue
		}
		if got := out.Bytes(); !reflect.DeepEqual(want, got) {
			t.Errorf("%s: want %q, got %q", f, want, got)
			t.Errorf("bson: %# x, %#v", want, v)
		}
	}
}

var errTests = []struct {
	f   string
	err error
}{
	// "test40.bson", // panic
	{"test41.bson", errors.New("corrupt BSON reading utf8 string")},
	//{"test42.bson", errors.New("corrupt BSON reading utf8 string") },
	{"overflow1.bson", io.ErrUnexpectedEOF},
	{"overflow2.bson", errors.New("corrupt document: want f bytes, have e")},
	{"overflow3.bson", errors.New("corrupt document: want c bytes, have b")},
	{"overflow4.bson", errors.New("corrupt BSON reading utf8 string")},

	//{"test43.bson", errors.New("corrupt BSON reading utf8 string") },
	//{"test44.bson", errors.New("corrupt BSON reading utf8 string") },
	//{"test45.bson", errors.New("corrupt BSON reading utf8 string") },
	// {"test57.bson", errors.New("corrupt BSON reading utf8 string") },

}

func TestLibBSONError(t *testing.T) {
	for _, tt := range errTests {
		f := filepath.Join("testdata", tt.f)
		want, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatal(f, err)
		}
		// test Unmarshall first
		v := make(map[string]interface{})
		if err := Unmarshal(want, &v); !reflect.DeepEqual(err, tt.err) {
			//			t.Errorf("Unmarshal: expected %v, got %v", tt.err, err)
			//		continue
		}
		// test Decode
		d := NewDecoder(bytes.NewReader(want))
		v = make(map[string]interface{})
		if err := d.Decode(&v); !reflect.DeepEqual(err, tt.err) {
			t.Errorf("Decode: expected %v, got %v", tt.err, err)
		}
	}
}
