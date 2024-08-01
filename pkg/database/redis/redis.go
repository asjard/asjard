package redis

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
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/redis/go-redis/v9"
)

const (
	defaultClientname = "default"
)

type ClientManager struct {
	clients sync.Map
}

type ClientConn struct {
	name   string
	client *redis.Client
}

type Config struct {
	Clients map[string]ClientConnConfig `json:"clients"`
	Options Options                     `json:"options"`
}

type ClientConnConfig struct {
	Address  string  `json:"address"`
	UserName string  `json:"username"`
	Password string  `json:"password"`
	DB       int     `json:"db"`
	Options  Options `json:"options"`
}

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
	CAFile                string             `json:"caFile"`
	CertFile              string             `json:"certFile"`
	KeyFile               string             `json:"keyFile"`
	DisableIndentity      bool               `json:"disableIndentity"`
	IndentitySuffix       string             `json:"indentitySuffix"`
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
	clientManager = &ClientManager{}
	bootstrap.AddBootstrap(clientManager)
}

func WithClientName(clientName string) func(*ClientOptions) {
	return func(opt *ClientOptions) {
		opt.clientName = clientName
	}
}

func Client(opts ...Option) (*redis.Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := clientManager.clients.Load(options.clientName)
	if !ok {
		logger.Error("redis not found", "name", options.clientName)
		return nil, status.InternalServerError
	}
	client, ok := conn.(*ClientConn)
	if !ok {
		logger.Error("invalid redis client, must be *ClientConn", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError
	}
	return client.client, nil
}

func (m *ClientManager) Bootstrap() error {
	clients, err := m.loadAndWatch()
	if err != nil {
		return err
	}
	return m.newClients(clients)
}

func (m *ClientManager) Shutdown() {
	m.clients.Range(func(key, value any) bool {
		conn, ok := value.(*ClientConn)
		if ok {
			if err := conn.client.Close(); err != nil {
				logger.Error("close redis client fail", "err", err)
			}
			m.clients.Delete(key)
		}
		return true
	})
}

func (m *ClientManager) newClients(clients map[string]*ClientConnConfig) error {
	logger.Debug("new clients", "conf", clients)
	for name, conf := range clients {
		logger.Debug("connect to redis", "name", name, "conf", conf)
		if err := m.newClient(name, conf); err != nil {
			return err
		}
	}
	return nil
}

func (m *ClientManager) newClient(name string, conf *ClientConnConfig) error {
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
	if conf.Options.CAFile != "" && conf.Options.CertFile != "" && conf.Options.KeyFile != "" {
		conf.Options.CAFile = filepath.Join(utils.GetCertDir(), conf.Options.CAFile)
		if !utils.IsPathExists(conf.Options.CAFile) {
			return fmt.Errorf("cafile %s not found", conf.Options.CAFile)
		}
		conf.Options.CertFile = filepath.Join(utils.GetCertDir(), conf.Options.CertFile)
		if !utils.IsPathExists(conf.Options.CertFile) {
			return fmt.Errorf("certfile %s not found", conf.Options.CertFile)
		}
		conf.Options.KeyFile = filepath.Join(utils.GetCertDir(), conf.Options.KeyFile)
		if !utils.IsPathExists(conf.Options.KeyFile) {
			return fmt.Errorf("keyfile %s not found", conf.Options.KeyFile)
		}
		caData, err := os.ReadFile(conf.Options.CAFile)
		if err != nil {
			return err
		}
		cert, err := tls.LoadX509KeyPair(conf.Options.CertFile, conf.Options.KeyFile)
		if err != nil {
			return err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		clientOptions.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      pool,
		}
	}
	client := redis.NewClient(clientOptions)
	ctx, cancel := context.WithTimeout(context.Background(), conf.Options.DialTimeout.Duration)
	defer cancel()
	if status := client.Ping(ctx); status.Err() != nil {
		return status.Err()
	}
	m.clients.Store(name, &ClientConn{
		name:   name,
		client: client,
	})
	return nil
}

func (m *ClientManager) loadAndWatch() (map[string]*ClientConnConfig, error) {
	clients, err := m.loadConfig()
	if err != nil {
		return nil, err
	}
	config.AddPatternListener("asjard.database.redis.*", m.watch)
	return clients, nil
}

func (m *ClientManager) loadConfig() (map[string]*ClientConnConfig, error) {
	clients := make(map[string]*ClientConnConfig)
	options := defaultOptions
	if err := config.GetWithUnmarshal("asjard.database.redis.options", &options); err != nil {
		return clients, err
	}
	if err := config.GetWithUnmarshal("asjard.database.redis.clients", &clients); err != nil {
		return clients, err
	}
	for name, client := range clients {
		client.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.database.redis.clients.%s.options", name), &client.Options); err != nil {
			return clients, err
		}
	}
	return clients, nil
}

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
		clientName: defaultClientname,
	}
}
