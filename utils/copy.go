package utils

import (
	"sync"
	"unsafe"
)

var bufPool = sync.Pool{New: func() any { return make([]byte, 0, 128) }}

// UnsafeString2Byte string类型转bytes类型
func UnsafeString2Byte(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UnsafeByte2String byte类型转string类型
func UnsafeByte2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func SafeByte2String(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	buf := bufPool.Get().([]byte)
	buf = append(buf[:0], b...)
	s := string(buf)
	bufPool.Put(buf[:0])
	return s
}
