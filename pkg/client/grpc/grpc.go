package grpc

import (
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
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
	return c
}

// NewConn 获取服务连接
func (c Client) NewConn(target string, clientOpts *client.ClientOptions) (grpc.ClientConnInterface, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	serverName := u.Path
	var options []grpc.DialOption
	balanceName := c.balanceName
	if clientOpts.Balancer != nil && balancer.Get(clientOpts.Balancer.Name()) != nil {
		balancer.Register(clientOpts.Balancer)
		balanceName = clientOpts.Balancer.Name()
	}
	options = append(options, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, balanceName)))
	certFile := config.GetString(fmt.Sprintf("clients.grpc.%s.certFile", serverName),
		config.GetString("clients.grpc.certFile", ""))
	if certFile != "" {
		certFile = filepath.Join(utils.GetCertDir(), certFile)
	}
	if certFile != "" {
		creds, err := credentials.NewClientTLSFromFile(certFile, serverName)
		if err != nil {
			return nil, err
		}
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	options = append(options, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time: config.GetDuration(fmt.Sprintf("clients.grpc.%s.options.keepalive.Time", serverName),
			config.GetDuration("clients.grpc.options.keepalive.Time", time.Second*20)),
		Timeout: config.GetDuration(fmt.Sprintf("client.grpc.%s.options.keepalive.Timeout", serverName),
			config.GetDuration("client.grpc.options.keepalive.Timeout", time.Second)),
		PermitWithoutStream: config.GetBool(fmt.Sprintf("client.grpc.%s.options.keepalive.PermitWithoutStream", serverName),
			config.GetBool("client.grpc.options.keepalive.PermitWithoutStream", true)),
	}))
	return grpc.NewClient(target, options...)
}
