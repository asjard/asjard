package utils

import "unsafe"

// String2Byte string类型转bytes类型
func String2Byte(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// Byte2String byte类型转string类型
func Byte2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}
