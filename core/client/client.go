package client

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"google.golang.org/grpc"
)

type ClientInterface interface {
	// target format asjard://grpc/serviceName
	NewConn(target string, options *ClientOptions) (grpc.ClientConnInterface, error)
}

// NewClientFunc 初始化客户端的方法
type NewClientFunc func(*ClientOptions) ClientInterface

var newClients = make(map[string]NewClientFunc)
var clients = make(map[string]ClientInterface)

// AddClient 添加客户端
func AddClient(protocol string, newClient NewClientFunc) {
	newClients[protocol] = newClient
}

// Init 客户端初始化
func Init() error {
	for protocol, newClient := range newClients {
		clients[protocol] = newClient(&ClientOptions{
			Resolver: &ClientBuilder{},
			Balancer: NewBalanceBuilder(config.GetString(fmt.Sprintf("clients.%s.loadbalances", protocol),
				config.GetString("clients.loadbalances", DefaultBalanceRoundRobin))),
			Interceptor: getChainUnaryInterceptors(config.GetString(fmt.Sprintf("clients.%s.interceptors", protocol),
				config.GetString("clients.interceptors", DefaultBalanceRoundRobin))),
		})
	}
	return nil
}

// Client 客户端
type Client struct {
	protocol   string
	serverName string
}

// Conn 链接地址
func (c Client) Conn() (grpc.ClientConnInterface, error) {
	cc, ok := clients[c.protocol]
	if !ok {
		return nil, fmt.Errorf("protocol %s client not found", c.protocol)
	}
	// 设置置指定服务的负载均衡
	options := &ClientOptions{
		Balancer:    NewBalanceBuilder(config.GetString(fmt.Sprintf("clients.%s.%s.loadbalances", c.protocol, c.serverName), "")),
		Interceptor: getChainUnaryInterceptors(config.GetString(fmt.Sprintf("clients.%s.%s.interceptors", c.protocol, c.serverName), "")),
	}
	return cc.NewConn(fmt.Sprintf("%s://%s/%s", constant.Framework, c.protocol, c.serverName), options)
}

// NewClient 新客户端
func NewClient(protocol, serverName string) *Client {
	return &Client{
		protocol:   protocol,
		serverName: serverName,
	}
}
