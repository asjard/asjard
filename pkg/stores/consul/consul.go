package consul

import (
	"fmt"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/initator"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/utils"
	"github.com/hashicorp/consul/api"
)

const (
	DefaultClientName = "default"
)

type ClientManager struct {
	clients sync.Map
}

type ClientConn struct {
	name   string
	client *api.Client
}

type Config struct {
	Clients map[string]ClientConnConfig `json:"clients"`
	Options Options                     `json:"options"`
}

type ClientConnConfig struct {
	Address    string             `json:"address"`
	Scheme     string             `json:"scheme"`
	PathPrefix string             `json:"pathPrefix"`
	Datacenter string             `json:"datacenter"`
	Username   string             `json:"username"`
	Password   string             `json:"password"`
	WaitTime   utils.JSONDuration `json:"waitTime"`
	Token      string             `json:"token"`
	Namespace  string             `json:"namespace"`
	Partition  string             `json:"partition"`
	Options    Options            `json:"options"`
}

type Options struct{}

type ClientOptions struct {
	clientName string
}

type Option func(*ClientOptions)

var (
	clientManager  *ClientManager
	defaultOptions = Options{}
)

func init() {
	clientManager = &ClientManager{}
	initator.AddInitator(clientManager)
}

func WithClientName(clientName string) Option {
	return func(opt *ClientOptions) {
		opt.clientName = clientName
	}
}

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

func (m *ClientManager) Start() error {
	clients, err := m.loadAndWatch()
	if err != nil {
		return err
	}
	return m.newClients(clients)
}

func (m *ClientManager) Stop() {}

func (m *ClientManager) newClients(clients map[string]*ClientConnConfig) error {
	for name, conf := range clients {
		logger.Debug("connect to consul", "name", name, "conf", conf)
		if err := m.newClient(name, conf); err != nil {
			return err
		}
	}
	return nil
}

func (m *ClientManager) newClient(name string, conf *ClientConnConfig) error {
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
	client, err := api.NewClient(apiConf)
	if err != nil {
		return err
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
	config.AddPatternListener("asjard.stores.consul", m.watch)
	return clients, nil
}

func (m *ClientManager) loadConfig() (map[string]*ClientConnConfig, error) {
	clients := make(map[string]*ClientConnConfig)
	options := defaultOptions
	if err := config.GetWithUnmarshal("asjard.stores.consul.options", &options); err != nil {
		return nil, err
	}
	if err := config.GetWithUnmarshal("asjard.stores.consul.clients", &clients); err != nil {
		return nil, err
	}
	for name, client := range clients {
		client.Options = options
		if err := config.GetWithUnmarshal(fmt.Sprintf("asjard.stores.consul.clients.%s.options", name), &client.Options); err != nil {
			return nil, err
		}
	}
	return clients, nil
}

func (m *ClientManager) watch(event *config.Event) {}

func defaultClientOptions() *ClientOptions {
	return &ClientOptions{
		clientName: DefaultClientName,
	}
}
