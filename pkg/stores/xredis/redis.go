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
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

const (
	// DefaultClientName 默认客户端名称
	DefaultClientName = "default"
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
	// 关闭连接检查,如果不关闭则在连接到redis后会发起ping请求
	DisableCheckStatus bool `json:"disableConnectCheck"`
	// 是否开启链路追踪
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
	clientManager = &ClientManager{}
	bootstrap.AddBootstrap(clientManager)
}

func WithClientName(clientName string) func(*ClientOptions) {
	return func(opt *ClientOptions) {
		if clientName != "" {
			opt.clientName = clientName
		}
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
		return nil, status.DatabaseNotFoundError()
	}
	client, ok := conn.(*ClientConn)
	if !ok {
		logger.Error("invalid redis client, must be *ClientConn", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError()
	}
	return client.client, nil
}

func (m *ClientManager) Start() error {
	clients, err := m.loadAndWatch()
	if err != nil {
		return err
	}
	return m.newClients(clients)
}

func (m *ClientManager) Stop() {
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
	logger.Debug("new clients", "clients", clients)
	for name, conf := range clients {
		if err := m.newClient(name, conf); err != nil {
			logger.Debug("connect to redis fail", "name", name, "conf", conf, "err", err)
			return err
		}
		logger.Debug("connect to redis success", "name", name, "conf", conf)
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
	if !conf.Options.DisableCheckStatus {
		ctx, cancel := context.WithTimeout(context.Background(), conf.Options.DialTimeout.Duration)
		defer cancel()
		if status := client.Ping(ctx); status.Err() != nil {
			return status.Err()
		}
	}
	if conf.Options.Traceable {
		if err := redisotel.InstrumentTracing(client, redisotel.WithDBSystem(name)); err != nil {
			return err
		}
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
	config.AddPatternListener("asjard.stores.redis.*", m.watch)
	return clients, nil
}

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
		clientName: DefaultClientName,
	}
}
