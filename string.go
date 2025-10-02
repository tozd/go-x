package x

import (
	"unsafe"
)

// String2ByteSlice converts a string to a byte slice without making a copy.
func String2ByteSlice(str string) []byte {
	return unsafe.Slice(unsafe.StringData(str), len(str)) //nolint:gosec
}

// ByteSlice2String converts a byte slice to a string without making a copy.
func ByteSlice2String(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(bs), len(bs)) //nolint:gosec
}
