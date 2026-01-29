/*
Package etcd implements the etcd configuration center integration,
fulfilling the core/config/Source interface.
*/
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
	// Name of the source.
	Name = "etcd"
	// Priority is 10, typically higher than local files but lower than CLI/ENV.
	Priority = 10

	defaultDelimiter = "/"
)

// Etcd manages the connection to etcd and the local cache of file-based configs.
type Etcd struct {
	options *config.SourceOptions
	app     runtime.APP
	conf    *Config
	client  *clientv3.Client
	// fileConfigs caches flattened properties from keys that are files (e.g., .yaml).
	fileConfigs map[string]map[string]any
	fcm         sync.RWMutex
}

// Config defines settings for the etcd source.
type Config struct {
	Client string `json:"client"` // Named etcd client from the stores package.
	// Delimiter for keys in etcd, defaults to "/".
	Delimiter string `json:"delimiter"`
}

var (
	defaultConfig = Config{
		Client:    xetcd.DefaultClientName,
		Delimiter: defaultDelimiter,
	}
)

func init() {
	// Register the etcd source with the global config manager.
	config.AddSource(Name, Priority, New)
}

// New initializes the etcd sourcer and starts the watch goroutines.
func New(options *config.SourceOptions) (config.Sourcer, error) {
	sourcer := &Etcd{
		app:         runtime.GetAPP(),
		fileConfigs: make(map[string]map[string]any),
		options:     options,
	}
	if err := sourcer.loadAndWatchConfig(); err != nil {
		return nil, err
	}
	return sourcer, nil
}

// GetAll performs an initial fetch of all keys under all relevant prefixes.
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
			// Process the KV; if it's a file, it returns multiple flattened values.
			for k, v := range s.getSetFileConfig(ref, key, priority, kv.Value) {
				result[k] = v
			}
		}
	}
	return result
}

// GetByKey fetches a single key directly from etcd.
func (s *Etcd) GetByKey(key string) any {
	for _, prefix := range s.prefixs() {
		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		// Convert internal dot key (a.b) to etcd path (a/b).
		resp, err := s.client.Get(ctx, prefix+strings.ReplaceAll(key, constant.ConfigDelimiter, s.conf.Delimiter))
		cancel()
		if err == nil && len(resp.Kvs) > 0 {
			return resp.Kvs[0].Value
		}
	}
	return nil
}

// Set is currently a placeholder for programmatic config updates.
func (s *Etcd) Set(key string, value any) error {
	return nil
}

func (s *Etcd) Priority() int { return Priority }
func (s *Etcd) Name() string  { return Name }
func (s *Etcd) Disconnect()   {}

// getSetFileConfig handles keys that represent files (e.g. "config.yaml").
// If the key has a supported extension, it parses the body and returns the properties.
func (s *Etcd) getSetFileConfig(file, key string, priority int, value []byte) map[string]*config.Value {
	ext := filepath.Ext(file)
	// If no extension, treat as a single literal value.
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
	// Flatten file content (YAML/JSON) into key-value pairs.
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
	logger.Error("convert to props map fail", "file", file, "err", err)
	return map[string]*config.Value{}
}

// getUpdateEvents determines what changed when an etcd key is updated.
// This is critical for keys containing entire files to detect which specific property changed.
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
	// Detect deleted properties within the file.
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
	// Detect created or updated properties within the file.
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

// loadAndWatchConfig sets up the etcd client and starts listening for changes.
func (s *Etcd) loadAndWatchConfig() error {
	if err := s.loadConfig(); err != nil {
		return err
	}
	// Allow the etcd client settings themselves to be updated dynamically.
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

// watch starts a background watch for every hierarchical prefix.
func (s *Etcd) watch() error {
	for priority, prefix := range s.prefixs() {
		go s.watchPrefix(prefix, priority)
	}
	return nil
}

// watchPrefix performs a long-running etcd Watch on a specific directory.
func (s *Etcd) watchPrefix(prefix string, priority int) {
	watchChan := s.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for resp := range watchChan {
		for _, event := range resp.Events {
			key := s.configKey(prefix, event.Kv.Key)
			ref := string(event.Kv.Key)
			switch event.Type {
			case mvccpb.PUT:
				// Push events to the framework for updates/creates.
				for _, event := range s.getUpdateEvents(ref, key, priority, event.Kv.Value) {
					s.options.Callback(event)
				}
			case mvccpb.DELETE:
				s.options.Callback(&config.Event{
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

// prefixs defines the search order for configs in etcd.
// Order: App-Global -> Env -> Service -> Region -> AZ -> Runtime(Instance)
func (s *Etcd) prefixs() []string {
	return []string{
		strings.Join([]string{s.prefix(), ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), "service", s.app.Instance.Group, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Group, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Group, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), "service", s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Group, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Group, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Group, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, s.app.Region, ""}, s.conf.Delimiter),
		strings.Join([]string{s.prefix(), s.app.Environment, "service", s.app.Instance.Name, s.app.Region, s.app.AZ, ""}, s.conf.Delimiter),

		strings.Join([]string{s.prefix(), "runtime", s.app.Instance.ID, ""}, s.conf.Delimiter),
	}
}

// prefix returns the base path: /{appName}/configs
func (s *Etcd) prefix() string {
	return strings.Join([]string{"", s.app.App, "configs"}, s.conf.Delimiter)
}

// configKey cleans the etcd key by removing the prefix and normalizing the delimiter.
func (s *Etcd) configKey(prefix string, key []byte) string {
	return strings.ReplaceAll(strings.TrimPrefix(string(key), prefix), s.conf.Delimiter, constant.ConfigDelimiter)
}
