package registry

import (
	"sync"

	"github.com/asjard/asjard/core/runtime"
)

// Options 服务注册发现相关参数
type Options struct {
	// 自定义服务选择key
	// 主要用来缓存
	// customePickFuncKeys []string
	pickFuncs []PickFunc
	watch     func(*Event)
	watchName string
}

// PickFunc 服务选择过滤方法
// 如果返回true则表示该实例满足要求
type PickFunc func(instance *Instance) bool

// Option .
type Option func(opts *Options)

var (
	// 默认参数
	defaultOptions     *Options
	defaultOptionsOnce sync.Once
)

// NewOptions .
func NewOptions(opts []Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// DefaultOptions 默认参数
// 相同的区域
// 相同的应用
// 相同的环境
func DefaultOptions() *Options {
	defaultOptionsOnce.Do(func() {
		app := runtime.GetAPP()
		defaultOptions = &Options{}
		opts := []Option{WithApp(app.App), WithRegion(app.Region), WithEnvironment(app.Environment)}
		for _, opt := range opts {
			opt(defaultOptions)
		}
	})
	return defaultOptions
}

// WithApp 设置APP名称
func WithApp(app string) func(opts *Options) {
	return func(opts *Options) {
		// opts.App = app
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.App == app
		})
	}
}

// WithPickFunc 自定义服务选择方法
func WithPickFunc(pickFuns []PickFunc) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, pickFuns...)
	}
}

// WithRegion 设置区域
func WithRegion(region string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Region == region
		})
	}
}

// WithEnvironment 设置环境
func WithEnvironment(environment string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Environment == environment
		})
	}
}

// WithServiceName 设置服务名称
func WithServiceName(serviceName string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Instance.Name == serviceName
		})
	}
}

func WithInstanceID(instanceID string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Instance.ID == instanceID
		})
	}
}

// WithRegistryName 设置注册/发现中心名称
func WithRegistryName(registryName string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.DiscoverName == registryName
		})
	}
}

// WithProtocol .
func WithProtocol(protocol string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			for protocol := range instance.Service.Endpoints {
				if protocol == protocol {
					return true
				}
			}
			return false
		})
	}
}

// WithVersion .
func WithVersion(version string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Instance.Version == version
		})
	}
}

// WithMetadata .
func WithMetadata(metadata map[string]string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			for wantKey, wantValue := range metadata {
				isOk := false
				for key, value := range instance.Service.Instance.MetaData {
					if wantKey == key && wantValue == value {
						isOk = true
						break
					}
				}
				if !isOk {
					return false
				}
			}
			return true
		})
	}
}

// WithWatch 更新服务变化
func WithWatch(watchName string, callback func(*Event)) func(opts *Options) {
	return func(opts *Options) {
		opts.watch = callback
		opts.watchName = watchName
	}
}

func okPickFunc() PickFunc {
	return func(instance *Instance) bool {
		return true
	}
}

func (opts *Options) getPickFuncs() []PickFunc {
	return opts.pickFuncs
}
