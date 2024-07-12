package grpc

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
)

const (
	// Protocol 协议名称
	Protocol = "grpc"
)

// Client .
type Client struct {
	balanceName string
	// 全局拦截器
	interceptor client.UnaryClientInterceptor
}

// ClientConn 客户端连接
type ClientConn struct {
	*grpc.ClientConn
	serviceName string
	protocol    string
}

func init() {
	client.AddClient(Protocol, NewClient)
}

// NewClient .
func NewClient(options *client.ClientOptions) client.ClientInterface {
	c := &Client{}
	if options.Resolver != nil {
		resolver.Register(options.Resolver)
	}
	if options.Balancer != nil {
		balancer.Register(options.Balancer)
		c.balanceName = options.Balancer.Name()
	}
	if options.Interceptor != nil {
		c.interceptor = options.Interceptor
	}
	return c
}

// ServiceName 客户端连接的服务名称
func (c ClientConn) ServiceName() string {
	return c.serviceName
}

// Protocol 客户端连接的协议
func (c ClientConn) Protocol() string {
	return c.protocol
}

// Conn .
func (c ClientConn) Conn() grpc.ClientConnInterface {
	return c.ClientConn
}

// NewConn 获取服务连接
// targe: ajard://grpc/{ServerName}
func (c Client) NewConn(target string, clientOpts *client.ClientOptions) (client.ClientConnInterface, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	serviceName := strings.Trim(u.Path, "/")
	var options []grpc.DialOption
	balanceName := c.balanceName
	if clientOpts.Balancer != nil && balancer.Get(clientOpts.Balancer.Name()) != nil {
		balancer.Register(clientOpts.Balancer)
		balanceName = clientOpts.Balancer.Name()
	}
	options = append(options, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, balanceName)))
	conf := serviceConfig(serviceName)
	if conf.CertFile != "" {
		conf.CertFile = filepath.Join(utils.GetCertDir(), conf.CertFile)
	}
	if conf.CertFile != "" {
		creds, err := credentials.NewClientTLSFromFile(conf.CertFile, serviceName)
		if err != nil {
			return nil, err
		}
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	options = append(options, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                conf.Options.Keepalive.Time.Duration,
		Timeout:             conf.Options.Keepalive.Timeout.Duration,
		PermitWithoutStream: conf.Options.Keepalive.PermitWithoutStream,
	}))
	interceptor := c.interceptor
	if clientOpts.Interceptor != nil {
		interceptor = clientOpts.Interceptor
	}
	if interceptor != nil {
		options = append(options,
			grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				return interceptor(ctx, method, req, reply, &ClientConn{ClientConn: cc, serviceName: serviceName, protocol: Protocol}, func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface) error {
					return invoker(ctx, method, req, reply, cc.Conn().(*grpc.ClientConn))
				})
			}))
	}
	conn, err := grpc.NewClient(target, options...)
	if err != nil {
		return nil, err
	}
	return &ClientConn{
		ClientConn:  conn,
		serviceName: serviceName,
		protocol:    Protocol,
	}, nil
}
