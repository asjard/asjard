package registry

import (
	"sync"

	"github.com/asjard/asjard/core/runtime"
)

// Options 服务注册发现相关参数
type Options struct {
	// 项目名称
	App string
	// 区域
	Region string
	// 环境
	Environment string
	// 服务名称
	ServiceName string
	// 服务注册发现中心名称
	RegistryName string
	// 协议名称
	Protocol string
	// 版本
	Version string
	// 服务元数据
	MetaData map[string]string

	// 自定义服务选择key
	// 主要用来缓存
	customePickFuncKeys []string
	pickFuncs           []PickFunc
	watch               func(*Event)
	watchName           string
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
		defaultOptions = &Options{
			App:         app.App,
			Region:      app.Region,
			Environment: app.Environment,
		}
	})
	return defaultOptions
}

// WithApp 设置APP名称
func WithApp(app string) func(opts *Options) {
	return func(opts *Options) {
		opts.App = app
		opts.pickFuncs = append(opts.pickFuncs, opts.appPickFunc())
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
		opts.Region = region
		opts.pickFuncs = append(opts.pickFuncs, opts.regionPickFunc())
	}
}

// WithEnvironment 设置环境
func WithEnvironment(environment string) func(opts *Options) {
	return func(opts *Options) {
		opts.Environment = environment
		opts.pickFuncs = append(opts.pickFuncs, opts.environmentPickFunc())
	}
}

// WithServiceName 设置服务名称
func WithServiceName(serviceName string) func(opts *Options) {
	return func(opts *Options) {
		opts.ServiceName = serviceName
		opts.pickFuncs = append(opts.pickFuncs, opts.servicePickFunc())
	}
}

// WithRegistryName 设置注册/发现中心名称
func WithRegistryName(registryName string) func(opts *Options) {
	return func(opts *Options) {
		opts.RegistryName = registryName
		opts.pickFuncs = append(opts.pickFuncs, opts.registryPickFunc())
	}
}

// WithProtocol .
func WithProtocol(protocol string) func(opts *Options) {
	return func(opts *Options) {
		opts.Protocol = protocol
		opts.pickFuncs = append(opts.pickFuncs, opts.protocolPickFunc())
	}
}

// WithVersion .
func WithVersion(version string) func(opts *Options) {
	return func(opts *Options) {
		opts.Version = version
		opts.pickFuncs = append(opts.pickFuncs, opts.versionPickFunc())
	}
}

// WithMetadata .
func WithMetadata(metadata map[string]string) func(opts *Options) {
	return func(opts *Options) {
		opts.MetaData = metadata
		opts.pickFuncs = append(opts.pickFuncs, opts.metadataPickFunc())
	}
}

// WithWatch 更新服务变化
func WithWatch(name string, callback func(*Event)) func(opts *Options) {
	return func(opts *Options) {
		opts.watch = callback
		opts.watchName = name
	}
}

func (opts *Options) okPickFunc() PickFunc {
	return func(instance *Instance) bool {
		return true
	}
}

// 应用选择
func (opts *Options) appPickFunc() PickFunc {
	if opts.App != "" {
		return func(instance *Instance) bool {
			return instance.Service.App == opts.App
		}
	}
	return opts.okPickFunc()
}

// 区域选择
func (opts *Options) regionPickFunc() PickFunc {
	if opts.Region != "" {
		return func(instance *Instance) bool {
			return instance.Service.Region == opts.Region
		}
	}
	return opts.okPickFunc()
}

// 区域选择
func (opts *Options) environmentPickFunc() PickFunc {
	if opts.Environment != "" {
		return func(instance *Instance) bool {
			return instance.Service.Environment == opts.Environment
		}
	}
	return opts.okPickFunc()
}

// 服务选择
func (opts *Options) servicePickFunc() PickFunc {
	if opts.ServiceName != "" {
		return func(instance *Instance) bool {
			return instance.Service.Instance.Name == opts.ServiceName
		}
	}
	return opts.okPickFunc()
}

// 服务发现中心选择
func (opts *Options) registryPickFunc() PickFunc {
	if opts.RegistryName != "" {
		return func(instance *Instance) bool {
			return instance.DiscoverName == opts.RegistryName
		}
	}
	return opts.okPickFunc()
}

// 协议选择
func (opts *Options) protocolPickFunc() PickFunc {
	if opts.Protocol != "" {
		return func(instance *Instance) bool {
			for protocol := range instance.Service.Endpoints {
				if protocol == opts.Protocol {
					return true
				}
			}
			return false
		}
	}
	return opts.okPickFunc()
}

// 版本号选择
func (opts *Options) versionPickFunc() PickFunc {
	if opts.Version != "" {
		return func(instance *Instance) bool {
			return instance.DiscoverName == opts.RegistryName
		}
	}
	return opts.okPickFunc()
}

// 元数据选择， 需要满足所有条件
func (opts *Options) metadataPickFunc() PickFunc {
	if opts.MetaData != nil && len(opts.MetaData) != 0 {
		return func(instance *Instance) bool {
			for wantKey, wantValue := range opts.MetaData {
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
		}
	}
	return opts.okPickFunc()
}

func (opts *Options) getPickFuncs() []PickFunc {
	return opts.pickFuncs
}
