package runtime

type Options struct {
	// 以分隔符开头
	startWithDelimiter bool
	// 以分隔符结束
	endWithDelimiter bool
	// 分隔符
	delimiter string
	// 不包含region信息
	withoutRegion bool
	// 不包含环境信息
	withoutEnv bool
	// 不包含service信息
	withoutService bool
	// 不包含版本信息
	withoutVersion bool
	// 包含serviceID
	withServiceId bool
}

type Option func(options *Options)

// WithStartWithDelimiter 以分隔符开头
func WithStartWithDelimiter(value bool) Option {
	return func(options *Options) {
		options.startWithDelimiter = value
	}
}

// WithEndWithDelimiter 以分隔符结尾
func WithEndWithDelimiter(value bool) Option {
	return func(options *Options) {
		options.endWithDelimiter = value
	}
}

// WithDelimiter 设置分隔符
func WithDelimiter(delimiter string) Option {
	return func(options *Options) {
		options.delimiter = delimiter
	}
}

// WithoutRegion 不包含region信息
func WithoutRegion(value bool) Option {
	return func(options *Options) {
		options.withoutRegion = value
	}
}

// WithoutEnv 不包含环境信息
func WithoutEnv(value bool) Option {
	return func(options *Options) {
		options.withoutEnv = value
	}
}

// WithoutService 不包含service信息
func WithoutService(value bool) Option {
	return func(options *Options) {
		options.withoutService = value
	}
}

// WithoutVersion 不包含版本信息
func WithoutVersion(value bool) Option {
	return func(options *Options) {
		options.withoutVersion = value
	}
}

// WithServiceId 用serviceId替换serviceName
func WithServiceId(value bool) Option {
	return func(options *Options) {
		options.withServiceId = value
	}
}

func defaultOptions() *Options {
	return &Options{
		delimiter: "/",
	}
}
