package validatepb

// Validater 参数校验需要实现的方法
type Validater interface {
	// 是否为有效的参数
	IsValid(options ...ValidaterOption) error
}

// ValidaterOptions 参数校验参数
type ValidaterOptions struct {
	FullMethod string
}

type ValidaterOption func(options *ValidaterOptions)

// WithFullMethod 设置方法名称
func WithFullMethod(fullMethod string) ValidaterOption {
	return func(options *ValidaterOptions) {
		options.FullMethod = fullMethod
	}
}
