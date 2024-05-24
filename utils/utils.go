package utils

// Uint32Len 获取整数长度
func Uint32Len(x uint32) uint32 {
	if x == 0 {
		return 1
	}
	var count uint32 = 0
	for x > 0 {
		x = x / 10
		count++
	}
	return count
}
