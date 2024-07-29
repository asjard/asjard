package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/database/etcd"
	"github.com/asjard/asjard/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// NAME 注册中心名称
	NAME = "etcd"
)

// Etcd etcd注册中心
type Etcd struct {
	client *clientv3.Client
	conf   *Config
}

// etcd配置
type Config struct {
	Client  string             `json:"client"`
	Timeout utils.JSONDuration `json:"timeout"`
}

var (
	_ registry.Register  = &Etcd{}
	_ registry.Discovery = &Etcd{}
	// 默认配置
	defaultConfig = Config{
		Client:  "default",
		Timeout: utils.JSONDuration{Duration: 5 * time.Second},
	}
	newEtcd *Etcd
	newOnce sync.Once
)

func init() {
	registry.AddRegister(NAME, NewRegister)
	registry.AddDiscover(NAME, NewDiscovery)
}

// New .
func NewRegister() (registry.Register, error) {
	return New()
}

func NewDiscovery() (registry.Discovery, error) {
	return New()
}

func New() (*Etcd, error) {
	var err error
	newOnce.Do(func() {
		etcdRegistry := &Etcd{}
		err = etcdRegistry.loadConfig()
		if err != nil {
			return
		}
		etcdRegistry.client, err = etcd.Client(etcd.WithClientName(etcdRegistry.conf.Client))
		if err != nil {
			return
		}
		newEtcd = etcdRegistry
	})
	if err != nil {
		return nil, err
	}
	return newEtcd, nil
}

// GetAll 获取服务实例
func (e *Etcd) GetAll() ([]*registry.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.conf.Timeout.Duration)
	defer cancel()
	resp, err := e.client.Get(ctx, e.prefix(), clientv3.WithPrefix())
	if err != nil {
		return []*registry.Instance{}, err
	}
	instances := make([]*registry.Instance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var service server.Service
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			return instances, err
		}
		instances = append(instances, &registry.Instance{
			DiscoverName: NAME,
			Service:      &service,
		})
	}
	return instances, nil
}

// Watch 监听服务变化
func (e *Etcd) Watch(callbak func(event *registry.Event)) {}

// HealthCheck 监控检查
func (e *Etcd) HealthCheck(instance *server.Service) error {
	return nil
}

// Name 名称
func (e *Etcd) Name() string {
	return NAME
}

// Registe 注册服务到注册中心
func (e *Etcd) Registe(instance *server.Service) error {
	logger.Debug("register instance into etcd", "instance", instance)
	b, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("register instance fail[%s]", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lease := clientv3.NewLease(e.client)
	grant, err := lease.Grant(ctx, 5)
	if err != nil {
		return err
	}
	if _, err := e.client.Put(ctx,
		e.registerKey(instance),
		string(b),
		clientv3.WithLease(grant.ID)); err != nil {
		return fmt.Errorf("register instance fail[%s]", err)
	}
	leaseChan, err := lease.KeepAlive(context.Background(), grant.ID)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case resp := <-leaseChan:
				if resp == nil {
					logger.Error("keepalive fail")
					for {
						logger.Debug("reregiste instance")
						if err := e.Registe(instance); err != nil {
							logger.Error("register instance fail", "err", err)
						} else {
							logger.Debug("reregiste instance success")
							return
						}
						time.Sleep(3 * time.Second)
					}
				}
			}
		}
	}()
	return nil
}

// Remove 从服务注册中心删除服务
func (e *Etcd) Remove(instance *server.Service) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if _, err := e.client.Delete(ctx, e.registerKey(instance)); err != nil {
		logger.Error("delete instance fail", "err", err)
	}
	cancel()
}

// Heartbeat 向服务注册中心发送心跳
func (e *Etcd) Heartbeat(instance *server.Service) {}

func (e *Etcd) loadConfig() error {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.registry.etcd", &conf); err != nil {
		return err
	}
	e.conf = &conf
	return nil
}

func (e *Etcd) registerKey(instance *server.Service) string {
	return fmt.Sprintf("%s/%s/%s", e.prefix(), instance.Instance.Name, instance.Instance.ID)
}

func (e *Etcd) prefix() string {
	return "/" + runtime.GetAPP().App + "/instances"
}
