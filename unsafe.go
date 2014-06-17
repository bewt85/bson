package bson

import "unsafe"

// These constants are evaluated at compile time. They are declared here to provide
// a way for implementations that do not allow unsafe to declare their replacement
// constants.

const (
	sizeofInt32 = int(unsafe.Sizeof(int32(1)))
	sizeofInt64 = int(unsafe.Sizeof(int64(1)))
	sizeofInt   = int(unsafe.Sizeof(int(1)))
	is64bit     = sizeofInt == sizeofInt64
)
