package xrabbitmq

import (
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
	"github.com/streadway/amqp"
)

const (
	// DefaultClientName 默认客户端名称
	DefaultClientName = "default"
)

type ClientManager struct {
	clients sync.Map

	cm      sync.RWMutex
	configs map[string]*ClientConnConfig
}

type ClientConn struct {
	name string
	conn *amqp.Connection
	conf *ClientConnConfig
	cm   sync.RWMutex
	done chan struct{}
}

type Config struct {
	Clients map[string]ClientConnConfig `json:"clients"`
	Options Options                     `json:"options"`
}

type ClientConnConfig struct {
	Url     string  `json:"url"`
	Vhost   string  `json:"vhost"`
	Options Options `json:"options"`
}

type Options struct {
	ChannelMax int                `json:"channelMax"`
	FrameSize  int                `json:"frameSize"`
	HeartBeat  utils.JSONDuration `json:"heartbeat"`
	CAFile     string             `json:"caFile"`
	CertFile   string             `json:"certFile"`
	KeyFile    string             `json:"keyFile"`
}

type ClientOptions struct {
	clientName string
}

type Option func(*ClientOptions)

var (
	clientManager  *ClientManager
	defaultOptions = Options{
		HeartBeat: utils.JSONDuration{Duration: time.Second},
	}
)

func init() {
	clientManager = &ClientManager{
		configs: make(map[string]*ClientConnConfig),
	}
	bootstrap.AddBootstrap(clientManager)
}

func WithClientName(clientName string) Option {
	return func(opt *ClientOptions) {
		if clientName != "" {
			opt.clientName = clientName
		}
	}
}

func (c *ClientConn) Channel() (*amqp.Channel, error) {
	c.cm.RLock()
	defer c.cm.RUnlock()
	if c.conn == nil || c.conn.IsClosed() {
		return nil, status.DatabaseNotFoundError()
	}
	return c.conn.Channel()
}

func Client(opts ...Option) (*ClientConn, error) {
	options := defaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}
	conn, ok := clientManager.clients.Load(options.clientName)
	if !ok {
		logger.Error("rabbitmq not found", "name", options.clientName)
		return nil, status.DatabaseNotFoundError()
	}
	client, ok := conn.(*ClientConn)
	if !ok {
		logger.Error("invalid rabbitmq client, must be *ClientConn", "current", fmt.Sprintf("%T", conn))
		return nil, status.InternalServerError()
	}
	return client, nil
}

func (c *ClientManager) Start() error {
	clients, err := c.loadAndWatch()
	if err != nil {
		return err
	}
	return c.newClients(clients)
}

func (c *ClientManager) Stop() {
	c.clients.Range(func(key, value any) bool {
		conn, ok := value.(*ClientConn)
		if ok && !conn.conn.IsClosed() {
			logger.Debug("rabbitmq close", "client", conn.name)
			// if err := conn.conn.Close(); err != nil {
			// 	logger.Error("close rabbitmq conn fail", "name", conn.name, "err", err)
			// }
			close(conn.done)
			c.clients.Delete(key)
		}
		return true
	})
}

func (c *ClientManager) newClients(clients map[string]*ClientConnConfig) error {
	logger.Debug("new clients", "clients", clients)
	for name, conf := range clients {
		conn, err := c.newClient(name, conf)
		if err != nil {
			logger.Error("connect to rabbitmq fail", "name", name, "conf", conf, "err", err)
			return err
		}
		clientConn := &ClientConn{
			name: name,
			conn: conn,
			conf: conf,
			done: make(chan struct{}),
		}
		v, ok := c.clients.Load(name)
		if ok {
			clientConn = v.(*ClientConn)
			clientConn.cm.Lock()
			clientConn.conn.Close()
			clientConn.name = name
			clientConn.conn = conn
			clientConn.conf = conf
			clientConn.cm.Unlock()
		} else {
			clientConn.keepalive()
		}
		c.clients.Store(name, clientConn)
		logger.Debug("connect to rabbitmq success", "name", name, "conf", conf)
	}
	return nil
}

func (c *ClientManager) newClient(name string, conf *ClientConnConfig) (*amqp.Connection, error) {
	dialConfig := amqp.Config{
		Vhost:      conf.Vhost,
		ChannelMax: conf.Options.ChannelMax,
		FrameSize:  conf.Options.FrameSize,
		Heartbeat:  conf.Options.HeartBeat.Duration,
	}
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
		dialConfig.TLSClientConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      pool,
		}
	}
	conn, err := amqp.DialConfig(conf.Url, dialConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *ClientManager) loadAndWatch() (map[string]*ClientConnConfig, error) {
	clients, err := c.loadConfig()
	if err != nil {
		return nil, err
	}
	config.AddPatternListener("asjard.stores.rabbitmq.*", c.watch)
	return clients, err
}

func (c *ClientManager) loadConfig() (map[string]*ClientConnConfig, error) {
	clients := make(map[string]*ClientConnConfig)
	options := defaultOptions
	if err := config.GetWithUnmarshal("asjard.stores.rabbitmq.options", &options); err != nil {
		return clients, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.rabbitmq.clients", &clients); err != nil {
		return clients, err
	}
	for name, client := range clients {
		client.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.stores.rabbitmq.clients.%s.options", name), &client.Options); err != nil {
			return clients, err
		}
	}
	c.cm.Lock()
	c.configs = clients
	c.cm.Unlock()
	return clients, nil
}

func (c *ClientManager) watch(event *config.Event) {
	clients, err := c.loadConfig()
	if err != nil {
		logger.Error("load rabbitmq config fail", "err", err)
		return
	}
	if err := c.newClients(clients); err != nil {
		logger.Error("new clients fail", "err", err)
		return
	}
	c.clients.Range(func(key, value any) bool {
		exist := false
		for clientName := range clients {
			if key.(string) == clientName {
				exist = true
				break
			}
		}
		if !exist {
			conn := value.(*ClientConn)
			close(conn.done)
			c.clients.Delete(key)
		}
		return true
	})
}

func (c *ClientConn) keepalive() {
	go func() {
		duration := time.Second
		for {
			select {
			case <-c.done:
				return
			default:
				if c.conn.IsClosed() {
					logger.Debug("rabbitmq disconnect, start to reconnect", "name", c.name)
					if err := clientManager.newClients(map[string]*ClientConnConfig{
						c.name: c.conf,
					}); err == nil {
						logger.Info("reconnect to rabbitmq success", "name", c.name)
						duration = time.Second
					} else {
						duration += time.Second
						if duration >= time.Second*10 {
							duration = time.Second * 10
						}
					}
				}
			}
			time.Sleep(duration)
		}
	}()
}

func defaultClientOptions() *ClientOptions {
	return &ClientOptions{
		clientName: DefaultClientName,
	}
}
