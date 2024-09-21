package validatepb

// Validater 参数校验需要实现的方法
type Validater interface {
	// 是否为有效的参数
	IsValid(fullMethod string) error
}
