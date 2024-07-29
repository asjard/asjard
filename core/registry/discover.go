package registry

const (
	// LocalRegistryName 本地注册发现中心名称
	LocalRegistryName = "local"
)

// Discovery 服务发现相关功能
type Discovery interface {
	// 获取所有服务实例
	GetAll() ([]*Instance, error)
	// 监听服务变化
	Watch(callbak func(event *Event))
	// 服务发现中心名称
	Name() string
}

// NewDiscoveryFunc 服务发现
type NewDiscoveryFunc func() (Discovery, error)

var (
	newDiscoverys = make(map[string]NewDiscoveryFunc)
)

// AddDiscover 添加服务发现组件
func AddDiscover(name string, newFunc NewDiscoveryFunc) error {
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
