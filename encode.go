package bson

import (
	"reflect"
	"unsafe"
)

// writer writes formatted BSON objects.
type writer struct {
	bson []byte
}

// writeMap encodes the contents of a map[string]interface{} as a BSON
// document.
func (w *writer) writeMap(v map[string]interface{}) error {
	off := len(w.bson)                  // the location of our header
	w.bson = append(w.bson, 0, 0, 0, 0) // document header
	var count int

	for k, v := range v {
		// write element key
		count += w.writeCstring(k)
		switch v := v.(type) {
		case int32:
			count += w.writeInt32(v)
		default:
			return &UnsupportedTypeError{reflect.TypeOf(v)}
		}

	}

	w.bson = append(w.bson, 0) // document trailer
	// update document header
	w.bson[off] = byte(count)
	w.bson[off+1] = byte(count >> 8)
	w.bson[off+2] = byte(count >> 16)
	w.bson[off+3] = byte(count >> 24)
	return nil
}

func (w *writer) writeCstring(s string) int {
	w.bson = append(w.bson, s...)
	w.bson = append(w.bson, 0)
	return len(s) + 1
}

func (w *writer) writeInt32(v int32) int {
	w.bson = append(w.bson,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24))
	return int(unsafe.Sizeof(v))
}

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "json: unsupported type: " + e.Type.String()
}
