package registry

import (
	"github.com/asjard/asjard/core/server"
)

// Register 服务注册相关功能
type Register interface {
	// 将服务注册到不同的注册中心
	Registe(service *server.Service) error
	// 从配置中心移除服务实例
	Remove(service *server.Service)
	// 注册中心名称
	Name() string
}

// NewRegisterFunc 用以启动配置中心方法
type NewRegisterFunc func() (Register, error)

// 注册的所有配置中心，会在启动阶段遍历并启动
var (
	newRegisters = make(map[string]NewRegisterFunc)
)

// AddRegister 添加服务注册组件
func AddRegister(name string, newFunc NewRegisterFunc) error {
	newRegisters[name] = newFunc
	return nil
}

// Registe 注册服务到注册中心
func Registe() error {
	return registryManager.registe()
}

// Unregiste 从注册中心删除服务
func Unregiste() error {
	return registryManager.remove()
}
