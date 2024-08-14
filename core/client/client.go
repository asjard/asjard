/*
Package client 不同协议的客户端维护，通过配置提供统一连接，拦截器的加载，负载均衡策略的选择
*/
package client

import (
	"fmt"
	"net/url"

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

// ConnOptions 连接参数
type ConnOptions struct {
	// 实例ID
	InstanceID string
	// 注册发现中心名称
	RegistryName string
}

type ConnOption func(opts *ConnOptions)

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
		interceptor, err := getChainUnaryInterceptors(protocol, conf)
		if err != nil {
			return err
		}
		clients[protocol] = newClient(&ClientOptions{
			Resolver:    &ClientBuilder{},
			Balancer:    NewBalanceBuilder(conf.Loadbalance),
			Interceptor: interceptor,
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

// NewClient 新客户端
func NewClient(protocol, serverName string) *Client {
	return &Client{
		protocol:   protocol,
		serverName: serverName,
	}
}

// Conn 链接地址
func (c Client) Conn(ops ...ConnOption) (grpc.ClientConnInterface, error) {
	cc, ok := clients[c.protocol]
	if !ok {
		return nil, fmt.Errorf("protocol %s client not found", c.protocol)
	}
	conf := serviceConfig(c.protocol, c.serverName)
	interceptor, err := getChainUnaryInterceptors(c.protocol, conf)
	if err != nil {
		return nil, err
	}
	// 设置置指定服务的负载均衡
	options := &ClientOptions{
		Balancer:    NewBalanceBuilder(conf.Loadbalance),
		Interceptor: interceptor,
	}
	return cc.NewConn(fmt.Sprintf("%s://%s/%s?%s",
		constant.Framework, c.protocol, c.serverName, c.connOptions(ops...).queryString()),
		options)
}

func (c Client) connOptions(ops ...ConnOption) *ConnOptions {
	options := &ConnOptions{}
	for _, op := range ops {
		op(options)
	}
	return options
}

func (o ConnOptions) queryString() string {
	v := make(url.Values)
	v.Set("instanceID", o.InstanceID)
	return v.Encode()

}

// WithInstanceID 客户端连接设置实例ID
func WithInstanceID(instanceID string) func(opts *ConnOptions) {
	return func(opts *ConnOptions) {
		opts.InstanceID = instanceID
	}
}

// WithRegistryName 客户端连接设置服务发现中心名称
func WithRegistryName(registryName string) func(opts *ConnOptions) {
	return func(opts *ConnOptions) {
		opts.RegistryName = registryName
	}
}
