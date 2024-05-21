package security

// Options 加解密参数
type Options struct {
	// 加解密组件名称
	cipherName string
	// 自定义参数
	params map[string]any
}

// Option .
type Option func(opts *Options)

// WithCipherName 设置加解密组件名称
func WithCipherName(name string) func(opts *Options) {
	return func(opts *Options) {
		if name != "" {
			opts.cipherName = name
		}
	}
}

// WithParams 设置定义参数
func WithParams(params map[string]any) func(opts *Options) {
	return func(opts *Options) {
		opts.params = params
	}
}

// Name 名称
func (opts *Options) Name() string {
	return opts.cipherName
}

// Params 自定义参数
func (opts *Options) Params() map[string]any {
	return opts.params
}
