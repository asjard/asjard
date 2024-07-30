package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/database/etcd"
	"github.com/asjard/asjard/utils"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// NAME 注册中心名称
	NAME = "etcd"
)

// Etcd etcd注册中心
type Etcd struct {
	client *clientv3.Client
	cb     func(event *registry.Event)
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
	discover, err := New()
	if err != nil {
		return discover, err
	}
	go discover.watch()
	return discover, nil
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
func (e *Etcd) Watch(callbak func(event *registry.Event)) {
	e.cb = callbak
}

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

func (e *Etcd) watch() {
	logger.Debug("watch instance from etcd")
	watchChan := e.client.Watch(context.Background(), e.prefix(), clientv3.WithPrefix())
	for resp := range watchChan {
		for _, event := range resp.Events {
			logger.Debug("instance from etcd updated",
				"key", string(event.Kv.Key),
				"event", event.Type.String())
			callbackEvent := &registry.Event{
				Instance: &registry.Instance{
					DiscoverName: e.Name(),
				},
			}
			switch event.Type {
			case mvccpb.PUT:
				callbackEvent.Type = registry.EventTypeUpdate
				var service server.Service
				if err := json.Unmarshal(event.Kv.Value, &service); err != nil {
					logger.Error("unmarshal service fail",
						"key", string(event.Kv.Key),
						"value", string(event.Kv.Value),
						"err", err)
					continue
				}
				callbackEvent.Instance.Service = &service
			case mvccpb.DELETE:
				callbackEvent.Type = registry.EventTypeDelete
				keyList := strings.Split(string(event.Kv.Key), "/")
				if len(keyList) > 0 {
					callbackEvent.Instance.Service = &server.Service{
						APP: runtime.APP{
							Instance: runtime.Instance{
								ID: keyList[len(keyList)-1],
							},
						},
					}
				}
			}
			if e.cb != nil {
				e.cb(callbackEvent)
			}
		}
	}
	logger.Debug("watch exit")
}

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
