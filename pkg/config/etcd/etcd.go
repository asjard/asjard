package etcd

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/database/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// Name 名称
	Name = "etcd"
	// Priority 优先级
	Priority = 10
)

// Etcd etcd配置
type Etcd struct {
	cb      func(*config.Event)
	configs map[string]*Value
	app     runtime.APP
	conf    *Config
	client  *clientv3.Client
}

type Value struct {
	priority int
	value    any
}

type Config struct {
	Client string `json:"client"`
}

var (
	defaultConfig = Config{
		Client: etcd.DefaultClientName,
	}
)

func init() {
	config.AddSource(Name, Priority, New)
}

// New 配置源初始化
func New() (config.Sourcer, error) {
	sourcer := &Etcd{
		app: runtime.GetAPP(),
	}
	err := sourcer.loadAndWatchConfig()
	if err != nil {
		return nil, err
	}

	sourcer.client, err = etcd.Client(etcd.WithClientName(sourcer.conf.Client))
	if err != nil {
		return nil, err
	}
	return sourcer, nil
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
func (s *Etcd) Disconnect() {
}

func (s *Etcd) loadAndWatchConfig() error {
	return nil
}

// /{app}/configs/global/
// /{app}/configs/{service}/
// /{app}/configs/{service}/{region}
// /{app}/configs/{service}/{region}/{az}
func (s *Etcd) prefixs() []string {
	return []string{
		fmt.Sprintf("%s/global/", s.prefix()),
		fmt.Sprintf("%s/%s/", s.prefix(), s.app.Instance.Name),
		fmt.Sprintf("%s/%s/%s/", s.prefix(), s.app.Instance.Name, s.app.Region),
		fmt.Sprintf("%s/%s/%s/%s/", s.prefix(), s.app.Instance.Name, s.app.Region, s.app.AZ),
	}
}

func (s *Etcd) prefix() string {
	return fmt.Sprintf("/%s/configs", s.app.App)
}
