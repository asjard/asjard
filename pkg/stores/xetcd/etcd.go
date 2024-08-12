package xetcd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/initator"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// DefaultClientName 默认客户端名称
	DefaultClientName = "default"
)

// ClientManager 客户端连接维护
type ClientManager struct {
	clients sync.Map
}

// ClientConn 客户端连接
type ClientConn struct {
	name   string
	client *clientv3.Client
}

// Config 客户端连接配置
type Config struct {
	Clients map[string]ClientConnConfig `json:"clients"`
	Options Options                     `json:"options"`
}

// Options 客户端连接参数
type Options struct {
	AutoSyncInterval      utils.JSONDuration `json:"autoSyncInterval"`
	DialTimeout           utils.JSONDuration `json:"dialTimeout"`
	DialKeepAliveTime     utils.JSONDuration `json:"dialKeepAliveTime"`
	DialKeepAliveTimeout  utils.JSONDuration `json:"dialKeepAliveTimeout"`
	MaxCallSendMsgSize    int                `json:"maxCallSendMsgSize"`
	MaxCallRecvMsgSize    int                `json:"maxCallRecvMsgSize"`
	UserName              string             `json:"userName"`
	Password              string             `json:"password"`
	RejectOldCluster      bool               `json:"rejectOldCluster"`
	PermitWithoutStream   bool               `json:"permitWithoutStream"`
	MaxUnaryRetries       uint               `json:"maxUnaryRetries"`
	BackoffWaitBetween    utils.JSONDuration `json:"backoffWaitBetween"`
	BackoffJitterFraction float64            `json:"backoffJitterFraction"`
	CAFile                string             `json:"caFile"`
	CertFile              string             `json:"CertFile"`
	KeyFile               string             `json:"keyFile"`
}

// 客户端连接配置
type ClientConnConfig struct {
	Endpoints utils.JSONStrings `json:"endpoints"`
	Options   Options           `json:"options"`
}

type ClientOptions struct {
	clientName string
}

type Option func(*ClientOptions)

var (
	clientManager  *ClientManager
	defaultOptions = Options{
		AutoSyncInterval:     utils.JSONDuration{},
		DialTimeout:          utils.JSONDuration{},
		DialKeepAliveTime:    utils.JSONDuration{},
		DialKeepAliveTimeout: utils.JSONDuration{},
		MaxCallSendMsgSize:   2 * 1024 * 1024,
		MaxCallRecvMsgSize:   math.MaxInt32,
	}
)

func init() {
	clientManager = &ClientManager{}
	initator.AddInitator(clientManager)
}

// WithClientName 设置客户端名称
func WithClientName(clientName string) func(*ClientOptions) {
	return func(opt *ClientOptions) {
		opt.clientName = clientName
	}
}

func Client(opts ...Option) (*clientv3.Client, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := clientManager.clients.Load(options.clientName)
	if !ok {
		logger.Error("etcd not found", "name", options.clientName)
		return nil, status.DatabaseNotFoundError()
	}
	client, ok := conn.(*ClientConn)
	if !ok {
		logger.Error("invalid etcd client, must be *ClientConn", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError()
	}
	return client.client, nil
}

func (m *ClientManager) Start() error {
	clients, err := m.loadAndWatchConfig()
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
				logger.Error("close etcd client fail", "client", conn.name, "err", err)
			}
			m.clients.Delete(key)
		}
		return true
	})
}

func (m *ClientManager) newClients(clients map[string]*ClientConnConfig) error {
	for name, cfg := range clients {
		logger.Debug("connect to etcd", "name", name, "cfg", cfg)
		if err := m.newClient(name, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (m *ClientManager) newClient(name string, cfg *ClientConnConfig) error {
	clientConfig := clientv3.Config{Endpoints: cfg.Endpoints,
		AutoSyncInterval:      cfg.Options.AutoSyncInterval.Duration,
		DialTimeout:           cfg.Options.DialTimeout.Duration,
		DialKeepAliveTime:     cfg.Options.DialKeepAliveTime.Duration,
		DialKeepAliveTimeout:  cfg.Options.DialKeepAliveTimeout.Duration,
		Username:              cfg.Options.UserName,
		Password:              cfg.Options.Password,
		RejectOldCluster:      cfg.Options.RejectOldCluster,
		PermitWithoutStream:   cfg.Options.PermitWithoutStream,
		MaxUnaryRetries:       cfg.Options.MaxUnaryRetries,
		BackoffWaitBetween:    cfg.Options.BackoffWaitBetween.Duration,
		BackoffJitterFraction: cfg.Options.BackoffJitterFraction,
	}
	if cfg.Options.CAFile != "" && cfg.Options.CertFile != "" && cfg.Options.KeyFile != "" {
		cfg.Options.CAFile = filepath.Join(utils.GetCertDir(), cfg.Options.CAFile)
		if !utils.IsPathExists(cfg.Options.CAFile) {
			return fmt.Errorf("cafile %s not found", cfg.Options.CAFile)
		}
		cfg.Options.CertFile = filepath.Join(utils.GetCertDir(), cfg.Options.CertFile)
		if !utils.IsPathExists(cfg.Options.CertFile) {
			return fmt.Errorf("certFile %s not found", cfg.Options.CertFile)
		}
		cfg.Options.KeyFile = filepath.Join(utils.GetCertDir(), cfg.Options.KeyFile)
		if !utils.IsPathExists(cfg.Options.KeyFile) {
			return fmt.Errorf("keyFile %s not found", cfg.Options.KeyFile)
		}
		cert, err := tls.LoadX509KeyPair(cfg.Options.CertFile, cfg.Options.KeyFile)
		if err != nil {
			return err
		}
		caData, err := os.ReadFile(cfg.Options.CAFile)
		if err != nil {
			return err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caData)
		clientConfig.TLS = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      pool,
		}
	}
	client, err := clientv3.New(clientConfig)
	if err != nil {
		return err
	}
	m.clients.Store(name, &ClientConn{
		name:   name,
		client: client,
	})
	return nil
}

func (m *ClientManager) loadAndWatchConfig() (map[string]*ClientConnConfig, error) {
	clients, err := m.loadConfig()
	if err != nil {
		return clients, err
	}
	config.AddPatternListener("asjard.stores.etcd.*", m.watch)
	return clients, nil
}

func (m *ClientManager) loadConfig() (map[string]*ClientConnConfig, error) {
	clients := make(map[string]*ClientConnConfig)
	options := defaultOptions
	if err := config.GetWithUnmarshal("asjard.stores.etcd.options", &options); err != nil {
		return clients, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.etcd.clients", &clients); err != nil {
		return clients, err
	}
	for name, client := range clients {
		client.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.stores.etcd.clients.%s.options", name), &client.Options); err != nil {
			return clients, err
		}
	}
	return clients, nil
}

func (m *ClientManager) watch(event *config.Event) {
	clients, err := m.loadConfig()
	if err != nil {
		logger.Error("load etcd config fail", "err", err)
		return
	}
	if err := m.newClients(clients); err != nil {
		logger.Error("new clients fail", "err", err)
		return
	}
	m.clients.Range(func(key, value any) bool {
		if _, ok := clients[key.(string)]; !ok {
			logger.Debug("delete etcd client", "client", key)
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
