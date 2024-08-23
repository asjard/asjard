package consul

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/stores/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

const (
	// Name 名称
	Name = "consul"
	// Priority 优先级
	Priority = 11

	defaultDelimiter = "/"
)

// Consul consul配置中心实现
type Consul struct {
	options *config.SourceOptions
	app     runtime.APP
	conf    *Config
	client  *api.Client
}

type Value struct {
	priority int
	value    any
}

type Config struct {
	Client    string `json:"client"`
	Delimiter string `json:"delimiter"`
}

var (
	defaultConfig = Config{
		Client:    consul.DefaultClientName,
		Delimiter: defaultDelimiter,
	}
)

func init() {
	config.AddSource(Name, Priority, New)
}

// New Consul配置中心初始化
func New(options *config.SourceOptions) (config.Sourcer, error) {
	sourcer := &Consul{
		app:     runtime.GetAPP(),
		options: options,
	}
	if err := sourcer.loadAndWatchConfig(); err != nil {
		return nil, err
	}
	return sourcer, nil
}

func (s *Consul) GetAll() map[string]*config.Value {
	return map[string]*config.Value{}
}

func (s *Consul) Set(key string, value any) error {
	return nil
}

func (s *Consul) Disconnect() {}

func (s *Consul) Priority() int {
	return Priority
}

func (s *Consul) Name() string {
	return Name
}

// /{app}/configs/
// /{app}/configs/{env}/
//
// /{app}/configs/service/{service}/
// /{app}/configs/service/{service}/{region}/
// /{app}/configs/service/{service}/{region}/{az}/
//
// /{app}/configs/{env}/service/{service}/
// /{app}/configs/{env}/service/{service}/{region}/
// /{app}/configs/{env}/service/{service}/{region}/{az}/
//
// /{app}/configs/runtime/{instance.ID}/
func (s *Consul) prefixs() []string {
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

func (s *Consul) prefix() string {
	return strings.Join([]string{s.app.App, "configs"}, s.conf.Delimiter)
}

func (s *Consul) configKey(prefix string, key string) string {
	return strings.ReplaceAll(strings.TrimPrefix(key, prefix), s.conf.Delimiter, constant.ConfigDelimiter)
}

func (s *Consul) loadAndWatchConfig() error {
	if err := s.loadConfig(); err != nil {
		return err
	}
	config.AddListener("asjard.config.consul.*", s.watchConfig)
	return s.watch()
}

func (s *Consul) loadConfig() error {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.config.consul", &conf); err != nil {
		logger.Error("load config fail", "err", err)
		return err
	}
	s.conf = &conf
	client, err := consul.Client(consul.WithClientName(s.conf.Client))
	if err != nil {
		logger.Error("new consul client fail", "err", err)
		return err
	}
	s.client = client
	return nil
}

func (s *Consul) watchConfig(event *config.Event) {
	if err := s.loadConfig(); err != nil {
		logger.Error("consul watch config fail", "err", err)
	}
}

func (s *Consul) watch() error {
	for priority, prefix := range s.prefixs() {
		if err := newConfigWatch(s, prefix, priority); err != nil {
			return err
		}
	}
	return nil
}

type configWatch struct {
	prefix   string
	priority int
	configs  map[string]uint64
	cm       sync.RWMutex
	s        *Consul
}

func newConfigWatch(s *Consul, prefix string, priority int) error {
	watcher := &configWatch{
		prefix:   prefix,
		priority: priority,
		configs:  make(map[string]uint64),
		s:        s,
	}
	pl, err := watch.Parse(map[string]any{
		"type":   "keyprefix",
		"prefix": prefix,
	})
	if err != nil {
		return err
	}
	pl.Handler = watcher.handler
	go func() {
		if err := pl.RunWithClientAndHclog(watcher.s.client, nil); err != nil {
			logger.Error("consul watch config with prefix fail", "prefix", prefix, "err", err)
		}
	}()
	return nil
}

func (w *configWatch) updateConfig(configs map[string]any, modifyIndex uint64) {
	w.cm.Lock()
	for key, value := range configs {
		if oldModifyIndex, ok := w.configs[key]; !ok || oldModifyIndex != modifyIndex {
			w.configs[key] = modifyIndex
			w.s.options.Callback(&config.Event{
				Type: config.EventTypeCreate,
				Key:  strings.TrimPrefix(key, w.prefix),
				Value: &config.Value{
					Sourcer:  w.s,
					Value:    value,
					Priority: w.priority,
				},
			})
		}
	}
	w.cm.Unlock()
}

func (w *configWatch) handler(_ uint64, data any) {
	switch d := data.(type) {
	case api.KVPairs:
		for _, kv := range d {
			ext := filepath.Ext(kv.Key)
			configs := map[string]any{kv.Key: kv.Value}
			var err error
			if ext != "" && config.IsExtSupport(ext) {
				configs, err = config.ConvertToProperties(ext, kv.Value)
				if err != nil {
					logger.Error("consul convert to props fail", "key", kv.Key, "err", err)
					continue
				}

			}
			w.updateConfig(configs, kv.ModifyIndex)
		}

		w.cm.Lock()
		for key := range w.configs {
			exist := false
			for _, kv := range d {
				ext := filepath.Ext(kv.Key)
				if ext != "" && config.IsExtSupport(ext) {
					configs, err := config.ConvertToProperties(ext, kv.Value)
					if err != nil {
						logger.Error("consul conver to props fail", "key", kv.Key, "err", err)
						continue
					}
					if _, ok := configs[key]; ok {
						exist = true
						break
					}
				} else {
					if kv.Key == key {
						exist = true
						break
					}
				}
			}
			if !exist {
				w.s.options.Callback(&config.Event{
					Type: config.EventTypeDelete,
					Key:  strings.TrimPrefix(key, w.prefix),
					Value: &config.Value{
						Sourcer:  w.s,
						Priority: w.priority,
					},
				})
				delete(w.configs, key)
			}
		}
		w.cm.Unlock()
	default:
		logger.Error("can not decide the watch type, must be api.KVPair", "data", data, "type", fmt.Sprintf("%T", data))
	}
}
