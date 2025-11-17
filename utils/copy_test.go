package utils

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestZeroCopy(t *testing.T) {
	t.Run("String2Byte", func(t *testing.T) {
		str := "123"
		b := UnsafeString2Byte(str)
		assert.Equal(t, unsafe.StringData(str), unsafe.SliceData(b))
		t.Logf("%v,%v, %s", unsafe.StringData(str), unsafe.SliceData(b), string(b))
	})
	t.Run("Byte2String", func(t *testing.T) {
		b := []byte("123")
		str := UnsafeByte2String(b)
		assert.Equal(t, unsafe.SliceData(b), unsafe.StringData(str))
	})
}

func BenchmarkZeroCopy(b *testing.B) {
	str := "helloStr"
	bt := []byte("helloByte")
	b.Run("String2Byte", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			UnsafeString2Byte(str)
		}
	})
	b.Run("String2Byte-Normal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = []byte(str)
		}
	})
	b.Run("Byte2String", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			UnsafeByte2String(bt)
		}
	})
	b.Run("Byte2String-Normal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = string(bt)
		}
	})
}
