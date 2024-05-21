package registry

import (
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/server"
)

// Register 服务注册相关功能
type Register interface {
	// 将服务注册到不同的配置中心
	Registe(instance *server.Instance) error
	// 从配置中心移除服务实例
	Remove(instance *server.Instance)
	// 心跳，表示本服务还活着
	// 防止注册中心将服务实例删除
	Heartbeat(instance *server.Instance)
	// 注册中心名称
	Name() string
}

// NewRegisterFunc 用以启动配置中心方法
type NewRegisterFunc func() (Register, error)

// 注册的所有配置中心，会在启动阶段遍历并启动
var newRegisters []NewRegisterFunc

// AddRegister 添加服务注册组件
func AddRegister(newFunc NewRegisterFunc) error {
	newRegisters = append(newRegisters, newFunc)
	return nil
}

// Registe 注册服务到注册中心
func Registe() error {
	if !config.GetBool("registry.autoRegiste", true) {
		return nil
	}
	return registryManager.registe()
}

// Unregiste 从注册中心删除服务
func Unregiste() error {
	if !config.GetBool("registry.autoRegiste", true) {
		return nil
	}
	return registryManager.remove()
}
