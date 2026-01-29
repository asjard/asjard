/*
Package consul implements the consul configuration center integration,
fulfilling the core/config/Source interface.
*/
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
	// Name of the configuration source.
	Name = "consul"
	// Priority of this source. Higher than files, allowing remote overrides.
	Priority = 11

	defaultDelimiter = "/"
)

// Consul manages the connection to the Consul KV store and handles configuration synchronization.
type Consul struct {
	options *config.SourceOptions
	app     runtime.APP // Current application runtime information (App name, env, region, etc.)
	conf    *Config     // Internal configuration for the Consul client itself
	client  *api.Client // The actual Consul API client
}

// Config defines the settings for the Consul source implementation.
type Config struct {
	Client    string `json:"client"`    // The name of the consul client to use from stores
	Delimiter string `json:"delimiter"` // Path separator, usually "/"
}

var (
	defaultConfig = Config{
		Client:    consul.DefaultClientName,
		Delimiter: defaultDelimiter,
	}
)

func init() {
	// Register the source into the global configuration manager.
	config.AddSource(Name, Priority, New)
}

// New initializes the Consul configuration source and starts the background watchers.
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

// GetAll is a placeholder as Consul uses a push-based 'watch' mechanism.
func (s *Consul) GetAll() map[string]*config.Value {
	return map[string]*config.Value{}
}

// Set is a placeholder; currently, this source is read-only for the application.
func (s *Consul) Set(key string, value any) error {
	return nil
}

// Disconnect cleans up any resources if the source is closed.
func (s *Consul) Disconnect() {}

func (s *Consul) Priority() int {
	return Priority
}

func (s *Consul) Name() string {
	return Name
}

// prefixs generates a prioritized list of Consul KV paths to watch.
// This implements a "specific overrides general" strategy.
func (s *Consul) prefixs() []string {
	return []string{
		// 1. Global app configs
		strings.Join([]string{s.prefix(), ""}, s.conf.Delimiter),
		// 2. Environment specific configs (e.g., prod vs dev)
		strings.Join([]string{s.prefix(), s.app.Environment, ""}, s.conf.Delimiter),

		// 3. Service specific configs
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Group, ""}, s.conf.Delimiter),
		// 4. Regional service configs
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Group, s.app.Region, ""}, s.conf.Delimiter),
		// 5. Availability Zone specific configs
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Group, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		// 3. Service specific configs
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, ""}, s.conf.Delimiter),
		// 4. Regional service configs
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		// 5. Availability Zone specific configs
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		// 6. Env + Service specific combinations
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Group, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Group, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Group, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		// 6. Env + Service specific combinations
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		// 7. Instance specific runtime configs (Highest specificity)
		strings.Join([]string{s.prefix(), "runtime", s.app.Instance.ID, ""}, s.conf.Delimiter),
	}
}

// prefix generates the base path for this application's configurations in Consul.
func (s *Consul) prefix() string {
	return strings.Join([]string{s.app.App, "configs"}, s.conf.Delimiter)
}

// configKey converts a Consul KV path back into a standard internal framework config key.
func (s *Consul) configKey(prefix string, key string) string {
	return strings.ReplaceAll(strings.TrimPrefix(key, prefix), s.conf.Delimiter, constant.ConfigDelimiter)
}

// loadAndWatchConfig bootstraps the Consul connection and starts watching for changes.
func (s *Consul) loadAndWatchConfig() error {
	if err := s.loadConfig(); err != nil {
		return err
	}
	// Watch the consul client settings themselves for changes.
	config.AddListener("asjard.config.consul.*", s.watchConfig)
	return s.watch()
}

// loadConfig initializes the Consul client.
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

// watch creates a separate long-polling watch for every path defined in prefixs().
func (s *Consul) watch() error {
	for priority, prefix := range s.prefixs() {
		if err := newConfigWatch(s, prefix, priority); err != nil {
			return err
		}
	}
	return nil
}

// configWatch manages a single Consul key-prefix watch session.
type configWatch struct {
	prefix   string
	priority int
	configs  map[string]uint64 // Tracks ModifyIndex for each key to prevent redundant updates
	cm       sync.RWMutex
	s        *Consul
}

// newConfigWatch starts a background goroutine to monitor a specific path in Consul.
func newConfigWatch(s *Consul, prefix string, priority int) error {
	watcher := &configWatch{
		prefix:   prefix,
		priority: priority,
		configs:  make(map[string]uint64),
		s:        s,
	}
	// Configure Consul's native watch plan for 'keyprefix' type.
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

// updateConfig sends change events back to the framework's core configuration system.
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

// handler is called by the Consul SDK whenever data changes under the prefix.
func (w *configWatch) handler(_ uint64, data any) {
	switch d := data.(type) {
	case api.KVPairs:
		// 1. Handle Updates and New Keys
		for _, kv := range d {
			ext := filepath.Ext(kv.Key)
			configs := map[string]any{kv.Key: kv.Value}
			var err error
			// If the key has an extension like .yaml or .json, parse the content into properties.
			if ext != "" && config.IsExtSupport(ext) {
				configs, err = config.ConvertToProperties(ext, kv.Value)
				if err != nil {
					logger.Error("consul convert to props fail", "key", kv.Key, "err", err)
					continue
				}
			}
			w.updateConfig(configs, kv.ModifyIndex)
		}

		// 2. Handle Deletions
		// Check our local cache against the Consul data to find missing keys.
		w.cm.Lock()
		for key := range w.configs {
			exist := false
			for _, kv := range d {
				ext := filepath.Ext(kv.Key)
				if ext != "" && config.IsExtSupport(ext) {
					configs, err := config.ConvertToProperties(ext, kv.Value)
					if err == nil {
						if _, ok := configs[key]; ok {
							exist = true
							break
						}
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
