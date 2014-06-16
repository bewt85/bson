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
	doclen, buf := readInt32(data)
	if len(data) != doclen {
		return ErrTooShort
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	switch rv := rv.Elem(); rv.Kind() {
	case reflect.Struct:
		buf := buf[:len(buf)-1]
		return decodeStruct(buf, rv)
	case reflect.Map:
		buf := buf[:len(buf)-1]
		return decodeMap(buf, rv)
	default:
		return &InvalidUnmarshalError{rv.Type()}
	}
}

func decodeStruct(data []byte, v reflect.Value) error {
	return nil
}

func decodeMap(data []byte, v reflect.Value) error {
	return nil
}

// bsonIter is an interator over a BSON document.
type bsonIter struct {
	// source bson document, mutated during iteration
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

// Next advances the iterator to the next element in BSON document.
// The element is available via the Element method. It returns false
// when the end of the document is reached, or an error occurs.
// After Next() returns false, the Err method will return any error
// that occured during walking the document.
func (b *bsonIter) Next() bool {
	switch len(b.bson) {
	case 0:
		// we've read everything
		return false
	case 1:
		// error, there must be at least 2 bytes remaining to be
		// valid BSON
		b.err = errors.New("corrupt BSON, only 1 byte remains")
		return false
	}
	i := bytes.IndexByte(b.bson[1:], 0)
	if i < 0 {
		b.err = errors.New("corrupt BSON")
		return false
	}
	i += 2
	ename, rest := b.bson[:i], b.bson[i:]
	var element []byte
	switch typ := ename[0]; typ {
	case 0x01:
		// double
		if len(rest) < 8 {
			b.err = errors.New("corrupt BSON reading double")
			return false
		}
		element, rest = rest[:8], rest[8:]
	case 0x02:
		// UTF-8 string
		if len(rest) < 5 {
			b.err = errors.New("corrupt BSON reading utf8 string len")
			return false
		}
		var elen int
		elen, rest = readInt32(rest)
		if len(rest) < elen {
			b.err = errors.New("corrupt BSON reading utf8 string")
			return false
		}
		element = rest[:elen]
		rest = rest[elen:]
	case 0x04:
		// array (as BSON document)
		var elen int
		elen, _ = readInt32(rest)
		if len(rest) < elen {
			b.err = fmt.Errorf("corrupt document: want %x bytes, have %x", elen, len(rest))
			return false
		}
		element = rest[:elen]
		rest = rest[elen:]
	case 0x10:
		// int32
		if len(rest) < 4 {
			b.err = errors.New("corrupt BSON reading int32")
			return false
		}
		element, rest = rest[:4], rest[4:]
	case 0x12:
		// int64
		if len(rest) < 8 {
			b.err = errors.New("corrupt BSON reading int64")
			return false
		}
		element, rest = rest[:8], rest[8:]
	default:
		b.err = &InvalidBSONTypeError{typ}
		return false
	}
	b.bson, b.ename, b.element = rest, ename, element
	return true
}

// Err returns the first error that was encountered during iteration.
func (b *bsonIter) Err() error {
	return b.err
}

// Element returns the most recent element verified by a call to Next.
func (b *bsonIter) Element() (byte, []byte, []byte) {
	return b.ename[0], b.ename[1:], b.element
}

// An InvalidBSONTypeError describes an unhandled BSON document element type.
type InvalidBSONTypeError struct {
	Type byte
}

func (e *InvalidBSONTypeError) Error() string {
	return fmt.Sprintf("bson: unknown element type %x", e.Type)
}
