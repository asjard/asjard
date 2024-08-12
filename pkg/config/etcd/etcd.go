package etcd

import (
	"context"
	"strings"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/database/etcd"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// Name 名称
	Name = "etcd"
	// Priority 优先级
	Priority = 10

	defaultDelimiter = "/"
)

// Etcd etcd配置
type Etcd struct {
	cb     func(*config.Event)
	app    runtime.APP
	conf   *Config
	client *clientv3.Client
}

type Value struct {
	priority int
	value    any
}

type Config struct {
	Client string `json:"client"`
	// 分隔符
	Delimiter string `json:"delimiter"`
}

var (
	defaultConfig = Config{
		Client:    etcd.DefaultClientName,
		Delimiter: defaultDelimiter,
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
	if err := sourcer.watch(); err != nil {
		return nil, err
	}
	return sourcer, nil
}

// GetAll .
func (s *Etcd) GetAll() map[string]*config.Value {
	result := make(map[string]*config.Value)
	for priority, prefix := range s.prefixs() {
		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		resp, err := s.client.Get(ctx, prefix, clientv3.WithPrefix())
		cancel()
		if err != nil {
			logger.Error("get prefix values fail", "prefix", prefix, "err", err)
			continue
		}
		for _, kv := range resp.Kvs {
			result[s.configKey(prefix, kv.Key)] = &config.Value{
				Sourcer:  s,
				Value:    kv.Value,
				Priority: priority,
			}
		}
	}
	return result

}

// GetByKey .
func (s *Etcd) GetByKey(key string) any {
	for _, prefix := range s.prefixs() {
		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		resp, err := s.client.Get(ctx, prefix+strings.ReplaceAll(key, constant.ConfigDelimiter, s.conf.Delimiter))
		cancel()
		if err == nil && len(resp.Kvs) > 0 {
			return resp.Kvs[0].Value
		}
	}
	return nil
}

// Set 添加配置到etcd中
// TODO 添加在runtime命名空间下, 需要带过期时间参数
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
	conf, err := s.loadConfig()
	if err != nil {
		return err
	}
	s.conf = conf
	config.AddListener("asjard.config.etcd.*", s.watchConfig)
	return nil
}

func (s *Etcd) loadConfig() (*Config, error) {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.config.etcd", &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func (s *Etcd) watchConfig(event *config.Event) {
	conf, err := s.loadConfig()
	if err != nil {
		logger.Error("load config fail", "err")
		return
	}
	s.conf = conf
	client, err := etcd.Client(etcd.WithClientName(s.conf.Client))
	if err != nil {
		logger.Error("new etcd client fail", "err")
		return
	}
	s.client = client
}

func (s *Etcd) watch() error {
	for priority, prefix := range s.prefixs() {
		go s.watchPrefix(prefix, priority)
	}
	return nil
}

func (s *Etcd) watchPrefix(prefix string, priority int) {
	watchChan := s.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for resp := range watchChan {
		for _, event := range resp.Events {
			callbackEvent := &config.Event{
				Key: s.configKey(prefix, event.Kv.Key),
				Value: &config.Value{
					Sourcer:  s,
					Priority: priority,
				},
			}
			switch event.Type {
			case mvccpb.PUT:
				callbackEvent.Type = config.EventTypeUpdate
				callbackEvent.Value.Value = event.Kv.Value
			case mvccpb.DELETE:
				callbackEvent.Type = config.EventTypeDelete
			}
			if s.cb != nil {
				s.cb(callbackEvent)
			}
		}
	}
}

// /{app}/configs/global/
// /{app}/configs/service/{service}/
// /{app}/configs/service/{service}/{region}/
// /{app}/configs/service/{service}/{region}/{az}/
// /{app}/configs/service/{env}/{service}/
// /{app}/configs/service/{env}/{service}/{region}/
// /{app}/configs/service/{env}/{service}/{region}/{az}/
// /{app}/configs/runtime/{instance.ID}/
func (s *Etcd) prefixs() []string {
	return []string{
		s.globalPrefix(),
		strings.Join([]string{s.prefix(), s.app.Instance.Name, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), s.app.Environment, s.app.Instance.Name, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),
		s.runtimePrefix(),
	}
}

func (s *Etcd) prefix() string {
	return strings.Join([]string{"", s.app.App, "configs"}, s.conf.Delimiter)
}

func (s *Etcd) runtimePrefix() string {
	return strings.Join([]string{s.prefix(), "runtime", s.app.Instance.ID, ""}, s.conf.Delimiter)
}

func (s *Etcd) globalPrefix() string {
	return strings.Join([]string{s.prefix(), "global", ""}, s.conf.Delimiter)
}

func (s *Etcd) configKey(prefix string, key []byte) string {
	return strings.ReplaceAll(strings.TrimPrefix(string(key), prefix), s.conf.Delimiter, constant.ConfigDelimiter)
}
