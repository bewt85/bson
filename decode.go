package bson

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
)

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "bson: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Ptr {
		return "bson: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "bson: Unmarshal(nil " + e.Type.String() + ")"
}

// decode decodes data into v according to the rules detailed in Unmarshal.
func decode(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	switch rv := rv.Elem(); rv.Kind() {
	case reflect.Struct:
		return decodeStruct(data, rv)
	case reflect.Map:
		return decodeMap(data, rv)
	default:
		return &InvalidUnmarshalError{rv.Type()}
	}
}

func decodeStruct(data []byte, v reflect.Value) error {
	iter := reader{bson: data[4 : len(data)-1]}
	for iter.Next() {

	}
	return iter.Err()
}

func decodeMap(data []byte, v reflect.Value) error {
	iter := reader{bson: data[4 : len(data)-1]}
	for iter.Next() {
		typ, ename, element := iter.Element()
		kv := reflect.ValueOf(string(trimlast(ename)))
		switch typ {
		case 0x02:
			// utf-8 string
			vv := reflect.ValueOf(string(trimlast(element)))
			v.SetMapIndex(kv, vv)
		case 0x10:
			element := int32(element[0]) | int32(element[1])<<8 | int32(element[2])<<16 | int32(element[3])<<24
			vv := reflect.ValueOf(element)
			v.SetMapIndex(kv, vv)
		case 0x12:
			element := int64(element[0]) | int64(element[1])<<8 | int64(element[2])<<16 | int64(element[3])<<24 | int64(element[4])<<32 | int64(element[5]<<40) | int64(element[6]<<48) | int64(element[7]<<56)
			vv := reflect.ValueOf(element)
			v.SetMapIndex(kv, vv)
		default:
			return &InvalidBSONTypeError{typ}
		}
	}
	return iter.Err()
}

func trimlast(s []byte) []byte { return s[:len(s)-1] }

// reader is an iterator over a BSON document.
type reader struct {
	// source bson document, mutated during read.
	bson []byte

	// ename of the next element, ename[0] contains the
	// type of the element, tf. the smallest ename slice
	// is 2 bytes. (0x02, 0x00)
	ename []byte

	// bson element that has been parsed _but_not_decoded_. The type
	// of the element is stored in ename[0]
	element []byte

	// last error, if any
	err error
}

// Next advances the reader to the next element in BSON document.
// The element is available via the Element method. It returns false
// when the end of the document is reached, or an error occurs.
// After Next() returns false, the Err method will return any error
// that occured during walking the document.
func (r *reader) Next() bool {
	switch len(r.bson) {
	case 0:
		// we've read everything
		return false
	case 1:
		// error, there must be at least 2 bytes remaining to be
		// valid BSON
		r.err = errors.New("corrupt BSON, only 1 byte remains")
		return false
	}
	i := bytes.IndexByte(r.bson[1:], 0)
	if i < 0 {
		r.err = errors.New("corrupt BSON ename")
		return false
	}
	i += 2
	ename, rest := r.bson[:i], r.bson[i:]
	var element []byte
	switch typ := ename[0]; typ {
	case 0x01:
		// double
		if len(rest) < 8 {
			r.err = errors.New("corrupt BSON reading double")
			return false
		}
		element, rest = rest[:8], rest[8:]
	case 0x02:
		// UTF-8 string
		if len(rest) < 5 {
			r.err = errors.New("corrupt BSON reading utf8 string len")
			return false
		}
		var elen int
		elen, rest = readInt32(rest)
		if len(rest) < elen {
			r.err = errors.New("corrupt BSON reading utf8 string")
			return false
		}
		element = rest[:elen]
		rest = rest[elen:]
	case 0x3:
		// BSON document
		fallthrough
	case 0x04:
		// array (as BSON document)
		var elen int
		elen, _ = readInt32(rest)
		if len(rest) < elen {
			r.err = fmt.Errorf("corrupt document: want %x bytes, have %x", elen, len(rest))
			return false
		}
		element = rest[:elen]
		rest = rest[elen:]
	case 0x09:
		// UTC datetime
		// int64
		if len(rest) < 8 {
			r.err = errors.New("corrupt BSON reading utc datetime")
			return false
		}
		element, rest = rest[:8], rest[8:]
	case 0x0a:
		// null
	case 0x0b:
		// regex
		if len(rest) < 2 {
			// need at least two bytes for empty cstrings
			r.err = errors.New("corrupt BSON reading regex")
			return false
		}
		i := bytes.IndexByte(rest, 0)
		if i < 0 {
			r.err = errors.New("corrupt BSON regex 1")
			return false
		}
		i++
		j := bytes.IndexByte(rest[i+1:], 0)
		if j < 0 {
			r.err = errors.New("corrupt BSON regex 2")
			return false
		}
		j++
		element, rest = rest[:i+j+1], rest[i+j+1:]
	case 0x10:
		// int32
		if len(rest) < 4 {
			r.err = errors.New("corrupt BSON reading int32")
			return false
		}
		element, rest = rest[:4], rest[4:]
	case 0x12:
		// int64
		if len(rest) < 8 {
			r.err = errors.New("corrupt BSON reading int64")
			return false
		}
		element, rest = rest[:8], rest[8:]
	default:
		r.err = &InvalidBSONTypeError{typ}
		return false
	}
	r.bson, r.ename, r.element = rest, ename, element
	return true
}

// Err returns the first error that was encountered.
func (r *reader) Err() error {
	return r.err
}

// Element returns the most recent element read by a call to Next.
func (r *reader) Element() (byte, []byte, []byte) {
	return r.ename[0], r.ename[1:], r.element
}

// An InvalidBSONTypeError describes an unhandled BSON document element type.
type InvalidBSONTypeError struct {
	Type byte
}

func (e *InvalidBSONTypeError) Error() string {
	return fmt.Sprintf("bson: unknown element type %x", e.Type)
}

// readInt32 returns the value of the first 4 bytes of buf as a little endian
// int32. The remaining bytes are return as a convenience.
// If there is less than 4 bytes of data in buf, the function will panic.
func readInt32(buf []byte) (int, []byte) {
	v := int(buf[0]) | int(buf[1])<<8 | int(buf[2])<<16 | int(buf[3])<<24
	return v, buf[4:]
}

// readCstring returns a []byte representing the cstring value, including
// the trailing \0.
func readCstring(buf []byte) ([]byte, []byte, error) {
	switch i := bytes.IndexByte(buf, 0); i {
	case -1:
		return nil, nil, errors.New("bson: cstring missing \\0")
	default:
		i++
		return buf[:i], buf[i:], nil
	}
}
