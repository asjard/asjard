package xredis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultClientName is the alias used for the primary Redis connection if none specified.
	DefaultClientName = "default"
)

// ClientManager handles the registration, initialization, and hot-reloading of Redis clients.
type ClientManager struct {
	// clients stores active *ClientConn instances, indexed by name.
	clients sync.Map

	cm sync.RWMutex
	// configs caches the raw configuration for change detection during watching.
	configs map[string]*ClientConnConfig
}

// ClientConn wraps the redis.Client with its identification name.
type ClientConn struct {
	name   string
	client *redis.Client
}

// Config represents the schema for Redis settings in the configuration center.
type Config struct {
	Clients map[string]ClientConnConfig `json:"clients"`
	Options Options                     `json:"options"`
}

// ClientConnConfig defines connection parameters and security settings for a specific Redis instance.
type ClientConnConfig struct {
	Address  string `json:"address"`
	UserName string `json:"username"`
	Password string `json:"password"`
	// CipherName allows Address/User/Pass to be stored as encrypted strings.
	CipherName   string         `json:"cipherName"`
	CipherParams map[string]any `json:"cipherParams"`
	DB           int            `json:"db"`
	Options      Options        `json:"options"`
}

// Options contains connection pool tuning, timeouts, and security certificate paths.
type Options struct {
	ClientName            string             `json:"clientName"`
	Protocol              int                `json:"protocol"`
	MaxRetries            int                `json:"maxRetries"`
	MinRetryBackoff       utils.JSONDuration `json:"minRetryBackoff"`
	MaxRetryBackoff       utils.JSONDuration `json:"mmaxRetryBackoff"`
	DialTimeout           utils.JSONDuration `json:"dialTimout"`
	ReadTimeout           utils.JSONDuration `json:"readTimeout"`
	WriteTimeout          utils.JSONDuration `json:"writeTimeout"`
	ContextTimeoutEnabled bool               `json:"contextTimeoutEnabled"`
	PoolFIFO              bool               `json:"poolFIFO"`
	PoolSize              int                `json:"poolSize"`
	PoolTimeout           utils.JSONDuration `json:"poolTimeout"`
	MinIdleConns          int                `json:"minIdleConns"`
	MaxIdleConns          int                `json:"maxIdleConns"`
	MaxActiveConns        int                `json:"maxActiveConns"`
	ConnMaxIdleTime       utils.JSONDuration `json:"connMaxIdleTime"`
	ConnMaxLifeTime       utils.JSONDuration `json:"connMaxLifeTime"`
	// TLS/SSL Configuration files
	CAFile           string `json:"caFile"`
	CertFile         string `json:"certFile"`
	KeyFile          string `json:"keyFile"`
	DisableIndentity bool   `json:"disableIndentity"`
	IndentitySuffix  string `json:"indentitySuffix"`
	// DisableCheckStatus: If false, the manager pings Redis immediately after connecting.
	DisableCheckStatus bool `json:"disableConnectCheck"`
	// Traceable: When true, enables OpenTelemetry instrumentation for Redis commands.
	Traceable bool `json:"Traceable"`
}

type ClientOptions struct {
	clientName string
}

type Option func(*ClientOptions)

var (
	clientManager  *ClientManager
	defaultOptions = Options{
		DialTimeout: utils.JSONDuration{Duration: 3 * time.Second},
	}
)

func init() {
	clientManager = &ClientManager{configs: make(map[string]*ClientConnConfig)}
	// Register with bootstrap to ensure Redis is ready before business logic starts.
	bootstrap.AddBootstrap(clientManager)
}

// WithClientName functional option to specify which Redis instance to retrieve.
func WithClientName(clientName string) func(*ClientOptions) {
	return func(opt *ClientOptions) {
		if clientName != "" {
			opt.clientName = clientName
		}
	}
}

// Client retrieves a shared Redis client instance from the manager.
func Client(opts ...Option) (*redis.Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := clientManager.clients.Load(options.clientName)
	if !ok {
		logger.Error("redis not found", "name", options.clientName)
		return nil, status.DatabaseNotFoundError()
	}
	client, ok := conn.(*ClientConn)
	if !ok {
		logger.Error("invalid redis client, must be *ClientConn", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError()
	}
	return client.client, nil
}

// NewClient creates a fresh Redis client based on the current configuration registry.
func NewClient(opts ...Option) (*redis.Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	clientManager.cm.RLock()
	connConfig, ok := clientManager.configs[options.clientName]
	clientManager.cm.RUnlock()
	if !ok {
		logger.Error("redis not found", "name", options.clientName)
		return nil, status.DatabaseNotFoundError()
	}
	return clientManager.newClient(options.clientName, connConfig)
}

// Start loads the initial configuration and establishes all connections.
func (m *ClientManager) Start() error {
	clients, err := m.loadAndWatch()
	if err != nil {
		return err
	}
	return m.newClients(clients)
}

// Stop gracefully closes all active Redis connections.
func (m *ClientManager) Stop() {
	m.clients.Range(func(key, value any) bool {
		conn, ok := value.(*ClientConn)
		if ok {
			logger.Debug("redis close", "client", conn.name)
			if err := conn.client.Close(); err != nil {
				logger.Error("close redis client fail", "name", conn.name, "err", err)
			}
			m.clients.Delete(key)
		}
		return true
	})
}

// newClients populates the client registry from a map of configurations.
func (m *ClientManager) newClients(clients map[string]*ClientConnConfig) error {
	logger.Debug("new clients", "clients", clients)
	for name, conf := range clients {
		client, err := m.newClient(name, conf)
		if err != nil {
			logger.Error("connect to redis fail", "name", name, "conf", conf, "err", err)
			return err
		}
		m.clients.Store(name, &ClientConn{
			name:   name,
			client: client,
		})
		logger.Debug("connect to redis success", "name", name, "conf", conf)
	}
	return nil
}

// newClientOptions transforms framework config into native go-redis options.
// Handles decryption of credentials and TLS certificate loading.
func (m *ClientManager) newClientOptions(conf *ClientConnConfig) (*redis.Options, error) {
	clientOptions := &redis.Options{
		Addr:                  conf.Address,
		ClientName:            conf.Options.ClientName,
		Protocol:              conf.Options.Protocol,
		Username:              conf.UserName,
		Password:              conf.Password,
		DB:                    conf.DB,
		MaxRetries:            conf.Options.MaxRetries,
		MinRetryBackoff:       conf.Options.MinRetryBackoff.Duration,
		MaxRetryBackoff:       conf.Options.MaxRetryBackoff.Duration,
		DialTimeout:           conf.Options.DialTimeout.Duration,
		ReadTimeout:           conf.Options.ReadTimeout.Duration,
		WriteTimeout:          conf.Options.WriteTimeout.Duration,
		ContextTimeoutEnabled: conf.Options.ContextTimeoutEnabled,
		PoolFIFO:              conf.Options.PoolFIFO,
		PoolSize:              conf.Options.PoolSize,
		PoolTimeout:           conf.Options.PoolTimeout.Duration,
		MinIdleConns:          conf.Options.MinIdleConns,
		MaxIdleConns:          conf.Options.MaxIdleConns,
		MaxActiveConns:        conf.Options.MaxActiveConns,
		ConnMaxIdleTime:       conf.Options.ConnMaxIdleTime.Duration,
		ConnMaxLifetime:       conf.Options.ConnMaxLifeTime.Duration,
		DisableIndentity:      conf.Options.DisableIndentity,
		IdentitySuffix:        conf.Options.IndentitySuffix,
	}

	// Decrypt sensitive connection details if a cipher is specified.
	if conf.CipherName != "" {
		var err error
		cipherOptions := []security.Option{security.WithCipherName(conf.CipherName), security.WithParams(conf.CipherParams)}
		clientOptions.Username, err = security.Decrypt(conf.UserName, cipherOptions...)
		if err != nil {
			return nil, err
		}
		clientOptions.Password, err = security.Decrypt(conf.Password, cipherOptions...)
		if err != nil {
			return nil, err
		}
		clientOptions.Addr, err = security.Decrypt(conf.Address, cipherOptions...)
		if err != nil {
			return nil, err
		}
	}

	// Load TLS certificates for encrypted connections (mTLS support).
	if conf.Options.CAFile != "" && conf.Options.CertFile != "" && conf.Options.KeyFile != "" {
		conf.Options.CAFile = filepath.Join(utils.GetCertDir(), conf.Options.CAFile)
		if !utils.IsPathExists(conf.Options.CAFile) {
			return nil, fmt.Errorf("cafile %s not found", conf.Options.CAFile)
		}
		conf.Options.CertFile = filepath.Join(utils.GetCertDir(), conf.Options.CertFile)
		if !utils.IsPathExists(conf.Options.CertFile) {
			return nil, fmt.Errorf("certfile %s not found", conf.Options.CertFile)
		}
		conf.Options.KeyFile = filepath.Join(utils.GetCertDir(), conf.Options.KeyFile)
		if !utils.IsPathExists(conf.Options.KeyFile) {
			return nil, fmt.Errorf("keyfile %s not found", conf.Options.KeyFile)
		}
		caData, err := os.ReadFile(conf.Options.CAFile)
		if err != nil {
			return nil, err
		}
		cert, err := tls.LoadX509KeyPair(conf.Options.CertFile, conf.Options.KeyFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		clientOptions.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      pool,
		}
	}
	return clientOptions, nil
}

// newClient creates the Redis client and applies tracing instrumentation.
func (m *ClientManager) newClient(name string, conf *ClientConnConfig) (*redis.Client, error) {
	clientOptions, err := m.newClientOptions(conf)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(clientOptions)

	// Perform connectivity check if not explicitly disabled.
	if !conf.Options.DisableCheckStatus {
		ctx, cancel := context.WithTimeout(context.Background(), conf.Options.DialTimeout.Duration)
		defer cancel()
		if status := client.Ping(ctx); status.Err() != nil {
			return nil, status.Err()
		}
	}

	// Inject OpenTelemetry tracing to monitor Redis command performance.
	if conf.Options.Traceable {
		if err := redisotel.InstrumentTracing(client, redisotel.WithDBSystem(name)); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// loadAndWatch reads the configuration and registers a listener for hot-reloading.
func (m *ClientManager) loadAndWatch() (map[string]*ClientConnConfig, error) {
	clients, err := m.loadConfig()
	if err != nil {
		return nil, err
	}
	config.AddPatternListener("asjard.stores.redis.*", m.watch)
	return clients, nil
}

// loadConfig unmarshals global and per-client Redis settings from the config core.
func (m *ClientManager) loadConfig() (map[string]*ClientConnConfig, error) {
	clients := make(map[string]*ClientConnConfig)
	options := defaultOptions
	if err := config.GetWithUnmarshal("asjard.stores.redis.options", &options); err != nil {
		return clients, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.redis.clients", &clients); err != nil {
		return clients, err
	}
	for name, client := range clients {
		client.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.stores.redis.clients.%s.options", name), &client.Options); err != nil {
			return clients, err
		}
	}
	m.cm.Lock()
	m.configs = clients
	m.cm.Unlock()
	return clients, nil
}

// watch handles configuration change events by updating active connections or removing stale ones.
func (m *ClientManager) watch(event *config.Event) {
	clients, err := m.loadConfig()
	if err != nil {
		logger.Error("load redis config fail", "err", err)
		return
	}
	if err := m.newClients(clients); err != nil {
		logger.Error("new clients fail", "err", err)
		return
	}
	// Clean up connections that are no longer present in the updated configuration.
	m.clients.Range(func(key, value any) bool {
		exist := false
		for clientName := range clients {
			if key.(string) == clientName {
				exist = true
				break
			}
		}
		if !exist {
			m.clients.Delete(key)
		}
		return true
	})
}

func defaultClientOptions() *ClientOptions {
	return &ClientOptions{
		clientName: DefaultClientName,
	}
}
