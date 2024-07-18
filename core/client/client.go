package client

import (
	"fmt"

	"github.com/asjard/asjard/core/constant"
	"google.golang.org/grpc"
)

type ClientInterface interface {
	// target format asjard://grpc/serviceName
	NewConn(target string, options *ClientOptions) (ClientConnInterface, error)
}

// ClientConnInterface 客户端需要实现的接口
// 对grpc.ClientConnInterface扩展
type ClientConnInterface interface {
	grpc.ClientConnInterface
	// 客户端连接的服务名称
	ServiceName() string
	// 客户端连接的协议
	Protocol() string
	Conn() grpc.ClientConnInterface
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
		conf := GetConfigWithProtocol(protocol)
		clients[protocol] = newClient(&ClientOptions{
			Resolver:    &ClientBuilder{},
			Balancer:    NewBalanceBuilder(conf.Loadbalance),
			Interceptor: getChainUnaryInterceptors(conf),
		})
	}
	return nil
}

// Client 客户端
type Client struct {
	protocol   string
	serverName string
	conf       Config
}

// Conn 链接地址
func (c Client) Conn() (grpc.ClientConnInterface, error) {
	cc, ok := clients[c.protocol]
	if !ok {
		return nil, fmt.Errorf("protocol %s client not found", c.protocol)
	}
	conf := serviceConfig(c.protocol, c.serverName)
	// 设置置指定服务的负载均衡
	options := &ClientOptions{
		Balancer:    NewBalanceBuilder(conf.Loadbalance),
		Interceptor: getChainUnaryInterceptors(conf),
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
