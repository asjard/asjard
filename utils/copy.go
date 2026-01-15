package utils

import (
	"sync"
	"unsafe"
)

// bufPool is a reusable pool of byte slices to minimize allocations during
// frequent string conversions. Initial capacity is set to 128 bytes.
var bufPool = sync.Pool{New: func() any { return make([]byte, 0, 128) }}

// UnsafeString2Byte performs a zero-copy conversion from string to []byte.
// WARNING: The resulting byte slice must NOT be modified, as strings in Go
// are immutable. Modifying the slice will result in a runtime panic or
// undefined behavior.
func UnsafeString2Byte(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UnsafeByte2String performs a zero-copy conversion from []byte to string.
// This is extremely fast as it does not allocate new memory; it simply
// points the string header to the existing byte slice's underlying data.
func UnsafeByte2String(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// SafeByte2String converts a byte slice to a string using a buffer pool.
// While standard `string(b)` creates a copy, this method uses sync.Pool
// to reuse temporary buffers, which is useful in high-concurrency scenarios
// to reduce the frequency of heap allocations.
func SafeByte2String(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// Retrieve a temporary buffer from the pool.
	buf := bufPool.Get().([]byte)
	// Reset the buffer and copy the data.
	buf = append(buf[:0], b...)
	// Convert to string (this still performs a copy to ensure the string is immutable).
	s := string(buf)
	// Return the buffer to the pool for reuse.
	bufPool.Put(buf[:0])
	return s
}
