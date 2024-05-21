package etcd

import (
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/server"
)

const (
	// NAME 注册中心名称
	NAME = "etcd"
)

// Etcd etcd注册中心
type Etcd struct{}

var _ registry.Register = &Etcd{}

func init() {
	registry.AddRegister(New)
}

// New .
func New() (registry.Register, error) {
	return &Etcd{}, nil
}

// GetAll 获取服务实例
func (e *Etcd) GetAll() ([]*server.Instance, error) {
	return []*server.Instance{}, nil
}

// Watch 监听服务变化
func (e *Etcd) Watch(callbak func(event *registry.Event)) {}

// HealthCheck 监控检查
func (e *Etcd) HealthCheck(instance *server.Instance) error {
	return nil
}

// Name 名称
func (e *Etcd) Name() string {
	return NAME
}

// Registe 注册服务到注册中心
func (e *Etcd) Registe(instance *server.Instance) error {
	return nil
}

// Remove 从服务注册中心删除服务
func (e *Etcd) Remove(instance *server.Instance) {}

// Heartbeat 向服务注册中心发送心跳
func (e *Etcd) Heartbeat(instance *server.Instance) {}
