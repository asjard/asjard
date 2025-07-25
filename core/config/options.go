package config

import (
	"time"

	"github.com/asjard/asjard/core/constant"
)

// Unmarshaler 反序列化需要实现的方法
type Unmarshaler interface {
	Unmarshal(data []byte, v any) error
}

// Options 定义了配置读取过程中的一些参数
type Options struct {
	// 是否监听配置变化
	watch *watchOptions
	// 可以添加多个配置源
	// 在添加配置时将会根据指定的配置源更新到远程配置中心
	// 如果为空则添加配置到所有的配置源中
	sourceNames []string
	// 加解密组件名称
	// 从配置源获取到配置后加密或解密
	// 向配置源添加加密或者解密后的数据
	// 依赖security包中的加解密组件
	cipherName string
	// 是否加解密
	cipher bool
	// 禁用自动解密
	// 如果value值前缀包含'encrypted_cipherName:encryptedValue'则自动解密
	disableAutoDecryptValue bool
	// 时区, 转化为time.Time时有用
	location *time.Location
	// 分隔符, 字符串转换为列表时有用, 默认空白符
	delimiter   string
	unmarshaler Unmarshaler
	// 忽略大小写
	ignoreCase bool
	// value转换为大写
	toUpper bool
	// value转换为小写
	toLower bool
	// key列表，按顺序获取,如果获取不到则获取下一个
	keys []string
}

// ListenFunc 监听方法，如果返回true则调用回调函数
type ListenFunc func(*Event) bool

type watchOptions struct {
	// 正则匹配
	pattern string
	// 回调方法，当配置发生变化后通过此回调方法回调
	callback CallbackFunc
	f        ListenFunc
}

func (w *watchOptions) clone() *watchOptions {
	return &watchOptions{
		pattern:  w.pattern,
		callback: w.callback,
		f:        w.f,
	}
}

// Option .
type Option func(*Options)

// WithWatch 监听配置
func WithWatch(callback CallbackFunc) func(opts *Options) {
	return func(opts *Options) {
		opts.watch = &watchOptions{
			callback: callback,
		}
	}
}

// WithMatchWatch 匹配监听
func WithMatchWatch(pattern string, callback CallbackFunc) func(opts *Options) {
	return func(opts *Options) {
		if pattern == "" {
			return
		}
		opts.watch = &watchOptions{
			pattern:  pattern,
			callback: callback,
		}
	}
}

// WithPrefixWatch 前缀监听
func WithPrefixWatch(prefix string, callback CallbackFunc) Option {
	return func(opts *Options) {
		opts.watch = &watchOptions{
			pattern:  prefix + ".*",
			callback: callback,
		}
	}
}

// WithFuncWatch 方法监听
// func WithFuncWatch(f WatchFunc, callback func(*Event)) func(opts *Options) {
// 	return func(opts *Options) {
// 		opts.watch = &watchOptions{
// 			f:        f,
// 			callback: callback,
// 		}
// 	}
// }

// WithSource 添加多个配置源
func WithSource(sourceName string) func(opts *Options) {
	return func(opts *Options) {
		// 去重
		for _, name := range opts.sourceNames {
			if name == sourceName {
				return
			}
		}
		opts.sourceNames = append(opts.sourceNames, sourceName)
	}
}

// 禁用自动解密value值
func WithDisableAutoDecryptValue() func(opts *Options) {
	return func(opts *Options) {
		opts.disableAutoDecryptValue = true
	}
}

// WithCipher 加解密
func WithCipher(cipherName string) func(opts *Options) {
	return func(opts *Options) {
		opts.cipher = true
		opts.cipherName = cipherName
	}
}

// WithLocation 设置时区
func WithLocation(location *time.Location) func(opts *Options) {
	return func(opts *Options) {
		opts.location = location
	}
}

// WithUnmarshaler 设置反序列化
func WithUnmarshaler(unmarshaler Unmarshaler) func(opts *Options) {
	return func(opts *Options) {
		opts.unmarshaler = unmarshaler
	}
}

// WithDelimiter 设置分隔符
func WithDelimiter(delimiter string) func(opts *Options) {
	return func(opts *Options) {
		opts.delimiter = delimiter
	}
}

// WithIgnoreCase 设置大小写敏感
func WithIgnoreCase() func(opts *Options) {
	return func(opts *Options) {
		opts.ignoreCase = true
	}
}

// WithToUpper 设置转换为大写
func WithToUpper() func(opts *Options) {
	return func(opts *Options) {
		opts.toUpper = true
	}
}

// WithToLower 设置转换为小写
func WithToLower() func(opts *Options) {
	return func(opts *Options) {
		opts.toLower = true
	}
}

// WithChain 链式获取
func WithChain(keys []string) func(opts *Options) {
	return func(opts *Options) {
		opts.keys = keys
	}
}

// GetOptions 获取参数
func GetOptions(opts ...Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if options.delimiter == "" {
		options.delimiter = constant.DefaultDelimiter
	}
	return options
}
