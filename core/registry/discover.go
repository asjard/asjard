package registry

import (
	"github.com/asjard/asjard/core/server"
)

// Discovery 服务发现相关功能
type Discovery interface {
	// 获取所有服务实例
	GetAll() ([]*server.Instance, error)
	// 监听服务变化
	Watch(callbak func(event *Event))
	// 健康检查, 用以检测服务是否还在服务发现中心中
	// 和watch功能类似，如果服务发现中心没有watch能力
	// 可以通过健康检测判断服务是否还在服务发现中心中
	HealthCheck(instance *server.Instance) error
	// 服务发现中心名称
	Name() string
}

// NewDiscoveryFunc 服务发现
type NewDiscoveryFunc func() (Discovery, error)

var newDiscoverys []NewDiscoveryFunc

// AddDiscover 添加服务发现组件
func AddDiscover(newFunc NewDiscoveryFunc) error {
	newDiscoverys = append(newDiscoverys, newFunc)
	return nil
}

// Discover 从服务发现中心发现服务
// func Discover() error {
// 	// 自动从注册中心发现服务
// 	if config.GetBool("registry.autoDiscove", false) {
// 		return registryManager.discove()
// 	}
// 	return nil
// }

// PickServices 获取服务列表
// 从本地缓存中获取符合要求的服务实例
func PickServices(options *Options) []*server.Instance {
	return registryManager.pick(options)
}
