package bson

import (
	"errors"
	"reflect"
)

// encode encodes v according to the rules of Marshal into a BSON document.
func encode(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return nil, &MarshalerError{Type: rv.Type(), Err: errors.New("was nil")}
	}
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	var w writer
	switch rv.Kind() {
	case reflect.Map:
		err := w.writeMap(rv)
		return w.bson, err
	}
	return nil, &MarshalerError{Type: rv.Type(), Err: errors.New("unsupported")}
}

type MarshalerError struct {
	Type reflect.Type
	Err  error
}

func (e *MarshalerError) Error() string {
	return "json: error calling MarshalJSON for type " + e.Type.String() + ": " + e.Err.Error()
}

// writer writes formatted BSON objects.
type writer struct {
	bson []byte
}

// writeMap encodes the contents of a map[string]interface{} as a BSON
// document.
func (w *writer) writeMap(v reflect.Value) error {
	off := len(w.bson)                  // the location of our header
	w.bson = append(w.bson, 0, 0, 0, 0) // document header
	count := sizeofInt32 + 1            // header plus trailing 0x0
	keys := v.MapKeys()
	for _, k := range keys {
		// write element key
		switch v := v.MapIndex(k).Elem(); v.Kind() {
		case reflect.Int32:
			count += w.writeType(0x10)
			count += w.writeCstring(k.String())
			count += w.writeInt32(int32(v.Int()))
		case reflect.Int64:
			count += w.writeType(0x12)
			count += w.writeCstring(k.String())
			count += w.writeInt64(int64(v.Int()))
		default:
			return &UnsupportedTypeError{v.Type()}
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

func (w *writer) writeType(typ byte) int {
	w.bson = append(w.bson, typ)
	return 1
}

func (w *writer) writeCstring(s string) int {
	w.bson = append(w.bson, s...)
	w.bson = append(w.bson, 0)
	return len(s) + 1
}

func (w *writer) writeInt32(v int32) int {
	w.bson = append(w.bson, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
	return sizeofInt32
}

func (w *writer) writeInt64(v int64) int {
	w.bson = append(w.bson, byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
	return 8 // sizeofInt64
}

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "json: unsupported type: " + e.Type.String()
}
