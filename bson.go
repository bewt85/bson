// Package bson implements encoding and decoding of BSON objects as defined
// at http://bsonspec.org/spec.html. The mapping between BSON objects and Go
// values is described in the documentation for the Marshal and Unmarshal functions.
package bson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"unsafe"
)

var ErrTooShort = errors.New("bson document too short")

// Marshal returns the BSON encoding of v.
//
// Struct values encode as JSON objects. Each exported struct field becomes
// a member of the object unless
//  - the field's tag is "-", or
//  - the field is empty and its tag specifies the "omitempty" option.
//
// Map values encode as JSON objects. The map's key type must be string;
// the object keys are used directly as map keys.
func Marshal(v interface{}) ([]byte, error) {
	return nil, nil
}

// Unmarshal parses the BSON-encoded data and stores the result in the
// value pointed to by v.
//
// Unmarshal uses the inverse of the encodings that Marshal uses,
// allocating maps, slices, and pointers as necessary.
//
// Portions of data may be retained by the decoded result in v. Data should
// not be reused.
func Unmarshal(data []byte, v interface{}) error {
	if len(data) < 5 {
		return ErrTooShort
	}
	return decode(data, v)
}

// A Decoder reads and decodes BSON objects from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next BSON-encoded value from its input and stores it in
// the value pointed to by v.
//
// See the documentation for Unmarshal for details about the conversion of
// BSON into a Go value.
func (d *Decoder) Decode(v interface{}) error {
	var header [4]byte
	if _, err := io.ReadFull(d.r, header[:]); err != nil {
		return err
	}
	doclen := int32(binary.LittleEndian.Uint32(header[:]))
	doclen -= int32(unsafe.Sizeof(doclen))
	r := io.LimitReader(d.r, int64(doclen))
	buf := bytes.NewBuffer(header[:])
	n, err := io.Copy(buf, r)
	if err != nil {
		return err
	}
	if n != int64(doclen) {
		return io.ErrUnexpectedEOF
	}
	return Unmarshal(buf.Bytes(), v)
}

// An Encoder writes BSON objects to an output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the BSON encoding of v to the stream, followed by a
// newline character.
//
// See the documentation for Marshal for details about the conversion of Go
// values to BSON.
func (e *Encoder) Encode(v interface{}) error {
	return nil
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
