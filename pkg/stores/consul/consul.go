package consul

import (
	"fmt"
	"sync"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/hashicorp/consul/api"
)

const (
	// DefaultClientName is used when no specific client name is provided in the options.
	DefaultClientName = "default"
)

// ClientManager maintains a thread-safe registry of initialized Consul clients
// and their associated configurations.
type ClientManager struct {
	clients sync.Map // Map of name -> *ClientConn (active connections)

	cm      sync.RWMutex
	configs map[string]*ClientConnConfig // Original configurations for dynamic updates
}

// ClientConn wraps the official Consul API client with its registration name.
type ClientConn struct {
	name   string
	client *api.Client
}

// Config represents the root structure for Consul configuration in YAML/JSON.
type Config struct {
	Clients map[string]ClientConnConfig `json:"clients"`
	Options Options                     `json:"options"`
}

// ClientConnConfig contains all parameters needed to establish a connection to Consul,
// including security settings for encrypted credentials.
type ClientConnConfig struct {
	Address    string `json:"address"`
	Scheme     string `json:"scheme"`     // http or https
	PathPrefix string `json:"pathPrefix"` // API path prefix
	Datacenter string `json:"datacenter"`
	Username   string `json:"username"`
	Password   string `json:"password"`

	// CipherName and CipherParams allow for sensitive info (like passwords or tokens)
	// to be stored encrypted in the config files and decrypted at runtime.
	CipherName   string         `json:"cipherName"`
	CipherParams map[string]any `json:"cipherParams"`

	WaitTime  utils.JSONDuration `json:"waitTime"`
	Token     string             `json:"token"`     // ACL Token
	Namespace string             `json:"namespace"` // Enterprise feature
	Partition string             `json:"partition"` // Enterprise feature
	Options   Options            `json:"options"`
}

type Options struct{}

// ClientOptions used for the functional options pattern when requesting a client.
type ClientOptions struct {
	clientName string
}

type Option func(*ClientOptions)

var (
	clientManager  *ClientManager
	defaultOptions = Options{}
)

func init() {
	clientManager = &ClientManager{configs: make(map[string]*ClientConnConfig)}
	// Register with bootstrap to ensure Consul is ready before servers start.
	bootstrap.AddInitiator(clientManager)
}

// WithClientName specifies which named consul instance to retrieve.
func WithClientName(clientName string) Option {
	return func(opt *ClientOptions) {
		if clientName != "" {
			opt.clientName = clientName
		}
	}
}

// Client retrieves an existing, cached Consul client from the manager.
func Client(opts ...Option) (*api.Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := clientManager.clients.Load(options.clientName)
	if !ok {
		logger.Error("consul not found", "name", options.clientName)
		return nil, status.DatabaseNotFoundError()
	}
	client, ok := conn.(*ClientConn)
	if !ok {
		logger.Error("invalid consul client, must be *api.Client", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError()
	}
	return client.client, nil
}

// NewClient creates a fresh Consul client instance based on the current configuration.
func NewClient(opts ...Option) (*api.Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	clientManager.cm.RLock()
	connConf, ok := clientManager.configs[options.clientName]
	clientManager.cm.RUnlock()
	if !ok {
		logger.Error("consul not found", "name", options.clientName)
		return nil, status.DatabaseNotFoundError()
	}
	return clientManager.newClient(options.clientName, connConf)
}

// Start triggers the initial configuration load and starts watching for config changes.
func (m *ClientManager) Start() error {
	clients, err := m.loadAndWatch()
	if err != nil {
		return err
	}
	return m.newClients(clients)
}

func (m *ClientManager) Stop() {}

// newClients initializes multiple clients and stores them in the registry.
func (m *ClientManager) newClients(clients map[string]*ClientConnConfig) error {
	for name, conf := range clients {
		logger.Debug("connect to consul", "name", name, "conf", conf)
		client, err := m.newClient(name, conf)
		if err != nil {
			return err
		}
		m.clients.Store(name, &ClientConn{
			name:   name,
			client: client,
		})
	}
	return nil
}

// newApiConfig converts the Asjard internal config to the official HashiCorp api.Config.
// It automatically handles decryption for Address, Username, and Password if a Cipher is provided.
func (m *ClientManager) newApiConfig(conf *ClientConnConfig) (*api.Config, error) {
	apiConf := &api.Config{
		Address:    conf.Address,
		Scheme:     conf.Scheme,
		PathPrefix: conf.PathPrefix,
		Datacenter: conf.Datacenter,
		WaitTime:   conf.WaitTime.Duration,
		Token:      conf.Token,
		Namespace:  conf.Namespace,
		Partition:  conf.Partition,
	}
	if conf.Username != "" && conf.Password != "" {
		apiConf.HttpAuth = &api.HttpBasicAuth{
			Username: conf.Username,
			Password: conf.Password,
		}
	}

	// Handle secure decryption of sensitive fields.
	if conf.CipherName != "" {
		var err error
		cipherOptions := []security.Option{security.WithCipherName(conf.CipherName), security.WithParams(conf.CipherParams)}

		apiConf.Address, err = security.Decrypt(conf.Address, cipherOptions...)
		if err != nil {
			return nil, err
		}

		if conf.Username != "" && conf.Password != "" {
			username, err := security.Decrypt(conf.Username, cipherOptions...)
			if err != nil {
				return nil, err
			}
			password, err := security.Decrypt(conf.Password, cipherOptions...)
			if err != nil {
				return nil, err
			}
			apiConf.HttpAuth = &api.HttpBasicAuth{Username: username, Password: password}
		}
	}
	return apiConf, nil
}

// newClient creates the official API client and validates connectivity by checking for a Leader.
func (m *ClientManager) newClient(name string, conf *ClientConnConfig) (*api.Client, error) {
	apiConf, err := m.newApiConfig(conf)
	if err != nil {
		return nil, err
	}
	client, err := api.NewClient(apiConf)
	if err != nil {
		return nil, err
	}
	// Connectivity check: ensure the cluster is alive.
	if _, err := client.Status().Leader(); err != nil {
		return nil, err
	}

	return client, nil
}

// loadAndWatch loads initial settings and registers a listener for real-time config updates.
func (m *ClientManager) loadAndWatch() (map[string]*ClientConnConfig, error) {
	clients, err := m.loadConfig()
	if err != nil {
		return nil, err
	}
	// Watch the specific consul configuration path for changes (e.g., address updates).
	config.AddPatternListener("asjard.stores.consul", m.watch)
	return clients, nil
}

// loadConfig fetches settings from the core configuration system and merges global options with client-specific ones.
func (m *ClientManager) loadConfig() (map[string]*ClientConnConfig, error) {
	clients := make(map[string]*ClientConnConfig)
	options := defaultOptions
	if err := config.GetWithUnmarshal("asjard.stores.consul.options", &options); err != nil {
		return nil, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.consul.clients", &clients); err != nil {
		return nil, err
	}

	// Merge global options into individual client configs if not overridden locally.
	for name, client := range clients {
		client.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.stores.consul.clients.%s.options", name), &client.Options); err != nil {
			return nil, err
		}
	}

	m.cm.Lock()
	m.configs = clients
	m.cm.Unlock()
	return clients, nil
}

// watch is the callback for configuration change events.
func (m *ClientManager) watch(event *config.Event) {
	clients, err := m.loadConfig()
	if err != nil {
		logger.Error("load consul config fail", "err", err)
		return
	}
	// Re-initialize clients with new settings.
	if err := m.newClients(clients); err != nil {
		logger.Error("new consul clients fail", "err", err)
	}
}

func defaultClientOptions() *ClientOptions {
	return &ClientOptions{
		clientName: DefaultClientName,
	}
}
