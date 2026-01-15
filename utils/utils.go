package utils

import "math"

// Uint32Len calculates the number of digits (length) in a uint32 integer.
// For example:
//   - Input: 0     => Output: 1
//   - Input: 100   => Output: 3
//   - Input: 42949 => Output: 5
func Uint32Len(n uint32) uint32 {
	if n == 0 {
		return 1
	}
	return uint32(math.Log10(float64(n))) + 1
}
