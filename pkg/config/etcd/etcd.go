package etcd

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/stores/xetcd"
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
	cb          func(*config.Event)
	app         runtime.APP
	conf        *Config
	client      *clientv3.Client
	fileConfigs map[string]map[string]any
	fcm         sync.RWMutex
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
		Client:    xetcd.DefaultClientName,
		Delimiter: defaultDelimiter,
	}
)

func init() {
	config.AddSource(Name, Priority, New)
}

// New 配置源初始化
func New() (config.Sourcer, error) {
	sourcer := &Etcd{
		app:         runtime.GetAPP(),
		fileConfigs: make(map[string]map[string]any),
	}
	if err := sourcer.loadAndWatchConfig(); err != nil {
		return nil, err
	}
	return sourcer, nil
}

// GetAll 获取etcd中的所有配置
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
			key := s.configKey(prefix, kv.Key)
			ref := string(kv.Key)
			for k, v := range s.getSetFileConfig(ref, key, priority, kv.Value) {
				result[k] = v
			}
		}
	}
	return result

}

// GetByKey 根据key获取配置
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

// Watch 配置更新回调
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

func (s *Etcd) getSetFileConfig(file, key string, priority int, value []byte) map[string]*config.Value {
	ext := filepath.Ext(file)
	if ext == "" || !config.IsExtSupport(ext) {
		return map[string]*config.Value{
			key: {
				Sourcer:  s,
				Value:    value,
				Priority: priority,
				Ref:      file,
			},
		}
	}
	s.fcm.Lock()
	defer s.fcm.Unlock()
	if _, ok := s.fileConfigs[file]; !ok {
		s.fileConfigs[file] = make(map[string]any)
	}
	propsMap, err := config.ConvertToProperties(ext, value)
	if err == nil {
		configMap := make(map[string]*config.Value, len(propsMap))
		for k, v := range propsMap {
			s.fileConfigs[file][k] = v
			configMap[k] = &config.Value{
				Sourcer:  s,
				Value:    v,
				Priority: priority,
				Ref:      file,
			}
		}
		return configMap
	}
	logger.Error("conver to props map fail", "file", file, "err", err)
	return map[string]*config.Value{}
}

func (s *Etcd) getUpdateEvents(file, key string, priority int, value []byte) []*config.Event {
	var events []*config.Event
	ext := filepath.Ext(file)
	if ext == "" || !config.IsExtSupport(ext) {
		events = append(events, &config.Event{
			Type: config.EventTypeUpdate,
			Key:  key,
			Value: &config.Value{
				Sourcer:  s,
				Value:    value,
				Priority: priority,
				Ref:      file,
			},
		})
		return events
	}
	s.fcm.Lock()
	defer s.fcm.Unlock()
	if _, ok := s.fileConfigs[file]; !ok {
		s.fileConfigs[file] = map[string]any{}
	}
	propsMap, err := config.ConvertToProperties(ext, value)
	if err != nil {
		logger.Error("convert to props map fail", "file", file, "err", err)
		return events
	}
	for key := range s.fileConfigs[file] {
		if _, ok := propsMap[key]; !ok {
			events = append(events, &config.Event{
				Type: config.EventTypeDelete,
				Key:  key,
				Value: &config.Value{
					Sourcer:  s,
					Priority: priority,
				},
			})
			delete(s.fileConfigs[file], key)
		}
	}
	for k, v := range propsMap {
		if oldValue, ok := s.fileConfigs[file][k]; !ok || oldValue != v {
			events = append(events, &config.Event{
				Type: config.EventTypeUpdate,
				Key:  k,
				Value: &config.Value{
					Sourcer:  s,
					Value:    v,
					Priority: priority,
					Ref:      file,
				},
			})
			s.fileConfigs[file][k] = v
		}
	}
	return events
}

func (s *Etcd) loadAndWatchConfig() error {
	if err := s.loadConfig(); err != nil {
		return err
	}
	config.AddListener("asjard.config.etcd.*", s.watchConfig)
	return s.watch()
}

func (s *Etcd) loadConfig() error {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.config.etcd", &conf); err != nil {
		logger.Error("get etcd config fail", "err", err)
		return err
	}
	s.conf = &conf
	client, err := xetcd.Client(xetcd.WithClientName(s.conf.Client))
	if err != nil {
		logger.Error("new etcd client fail", "err", err)
		return err
	}
	s.client = client
	return nil
}

func (s *Etcd) watchConfig(event *config.Event) {
	s.loadConfig()
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
			key := s.configKey(prefix, event.Kv.Key)
			ref := string(event.Kv.Key)
			logger.Debug("etcd config event", "event", event.Type.String(), "key", key, "prefix", prefix)
			switch event.Type {
			case mvccpb.PUT:
				for _, event := range s.getUpdateEvents(ref, key, priority, event.Kv.Value) {
					s.cb(event)
				}
			case mvccpb.DELETE:
				s.cb(&config.Event{
					Type: config.EventTypeDelete,
					Key:  key,
					Value: &config.Value{
						Sourcer:  s,
						Ref:      ref,
						Priority: priority,
					},
				})
			}
		}
	}
}

// /{app}/configs/
// /{app}/configs/{env}/
//
// /{app}/configs/service/{service}/
// /{app}/configs/service/{service}/{region}/
// /{app}/configs/service/{service}/{region}/{az}/

// /{app}/configs/{env}/service/{service}/
// /{app}/configs/{env}/service/{service}/{region}/
// /{app}/configs/{env}/service/{service}/{region}/{az}/
//
// /{app}/configs/runtime/{instance.ID}/
// 以文件名后缀结尾的展开
func (s *Etcd) prefixs() []string {
	return []string{
		strings.Join([]string{s.prefix(), ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), "runtime", s.app.Instance.ID, ""}, s.conf.Delimiter),
	}
}

func (s *Etcd) prefix() string {
	return strings.Join([]string{"", s.app.App, "configs"}, s.conf.Delimiter)
}

func (s *Etcd) configKey(prefix string, key []byte) string {
	return strings.ReplaceAll(strings.TrimPrefix(string(key), prefix), s.conf.Delimiter, constant.ConfigDelimiter)
}
