package etcd

import "github.com/asjard/asjard/core/config"

const (
	// Name 名称
	Name = "etcd"
	// Priority 优先级
	Priority = 10
)

// Etcd etcd配置
type Etcd struct {
	cb func(*config.Event)
}

func init() {
	config.AddSource(Name, Priority, New)
}

// New .
func New() (config.Sourcer, error) {
	return &Etcd{}, nil
}

// GetAll .
func (s *Etcd) GetAll() map[string]*config.Value {
	return nil

}

// GetByKey .
func (s *Etcd) GetByKey(key string) any {
	return nil
}

// Set .
func (s *Etcd) Set(key string, value any) error {
	return nil
}

// Watch .
func (s *Etcd) Watch(cb func(*config.Event)) error {
	s.cb = cb
	return nil
}

// Priority 配置中心优先级
func (s *Etcd) Priority() int {
	return Priority
}

// Name 配置源名称
func (s *Etcd) Name() string {
	return Name
}

// DisConnect 断开连接
func (s *Etcd) DisConnect() {
}
