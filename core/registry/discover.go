package registry

import "github.com/asjard/asjard/core/logger"

const (
	// LocalRegistryName 本地注册发现中心名称
	LocalRegistryName = "local"
)

// Discovery 服务发现相关功能
type Discovery interface {
	// 获取所有服务实例
	GetAll() ([]*Instance, error)
	// 监听服务变化
	// Watch(callbak func(event *Event))
	// 服务发现中心名称
	Name() string
}

// CallbackFunc 回调方法
type CallbackFunc func(event *Event)

// DiscoveryOptions 服务发现初始化参数列表
type DiscoveryOptions struct {
	Callback CallbackFunc
}

// DiscoveryOption 服务发现初始化参数
type DiscoveryOption func(options *DiscoveryOptions)

// NewDiscoveryFunc 服务发现
type NewDiscoveryFunc func(options *DiscoveryOptions) (Discovery, error)

// WithDiscoveryCallback 设置服务发现回调函数
func WithDiscoveryCallback(callback CallbackFunc) func(options *DiscoveryOptions) {
	return func(options *DiscoveryOptions) {
		options.Callback = callback
	}
}

// NewDiscoveryOptions 服务发现参数初始化
func NewDiscoveryOptions(opts ...DiscoveryOption) *DiscoveryOptions {
	options := &DiscoveryOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

var (
	newDiscoverys = make(map[string]NewDiscoveryFunc)
)

// AddDiscover 添加服务发现组件
func AddDiscover(name string, newFunc NewDiscoveryFunc) error {
	logger.Debug("add discover", "name", name)
	newDiscoverys[name] = newFunc
	return nil
}

// Discover 服务发现
func Discover() error {
	return registryManager.discove()
}

// PickServices 获取服务列表
// 从本地缓存中获取符合要求的服务实例
func PickServices(opts ...Option) []*Instance {
	options := NewOptions(opts)
	if len(opts) == 0 {
		options = DefaultOptions()
	}
	return registryManager.pick(options)
}

// RemoveListener 移除监听
func RemoveListener(name string) {
	registryManager.removeListener(name)
}
