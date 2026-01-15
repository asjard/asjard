/*
Package grpc implements the built-in gRPC client, fulfilling the core/client/ClientInterface.
It handles connection management, service name resolution, and load balancing for gRPC traffic.
*/
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
	// Protocol defines the identifier for this client implementation.
	Protocol = "grpc"
)

// Client handles the creation of gRPC connections and maintains global settings
// like the default balancer and interceptors.
type Client struct {
	balanceName string
	// Global interceptor for all connections created by this client.
	interceptor client.UnaryClientInterceptor
}

// ClientConn wraps a standard grpc.ClientConn to satisfy the framework's ClientConnInterface.
type ClientConn struct {
	*grpc.ClientConn
	serviceName string
	protocol    string
}

func init() {
	// Automatically register the gRPC client factory into the framework's client manager.
	client.AddClient(Protocol, NewClient)
}

// NewClient initializes a Client instance and registers any custom resolvers or balancers provided.
func NewClient(options *client.ClientOptions) client.ClientInterface {
	c := &Client{}
	// Register custom naming resolver (e.g., ETCD, Consul) if provided.
	if options.Resolver != nil {
		resolver.Register(options.Resolver)
	}
	// Register custom load balancing strategy if provided.
	if options.Balancer != nil {
		balancer.Register(options.Balancer)
		c.balanceName = options.Balancer.Name()
	}
	if options.Interceptor != nil {
		c.interceptor = options.Interceptor
	}
	return c
}

// ServiceName returns the name of the target service for this connection.
func (c ClientConn) ServiceName() string {
	return c.serviceName
}

// Protocol returns "grpc".
func (c ClientConn) Protocol() string {
	return c.protocol
}

// Conn returns the underlying gRPC connection interface.
func (c ClientConn) Conn() grpc.ClientConnInterface {
	return c.ClientConn
}

// NewConn establishes a new gRPC client connection to a target.
// target format: asjard://grpc/{ServerName}
func (c Client) NewConn(target string, clientOpts *client.ClientOptions) (client.ClientConnInterface, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	serviceName := strings.Trim(u.Path, "/")
	var options []grpc.DialOption

	// 1. Configure Load Balancing
	balanceName := c.balanceName
	if clientOpts.Balancer != nil && balancer.Get(clientOpts.Balancer.Name()) == nil {
		balancer.Register(clientOpts.Balancer)
		balanceName = clientOpts.Balancer.Name()
	}
	options = append(options, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}]}`, balanceName)))

	// 2. Load gRPC specific configuration (Keepalive, TLS, etc.)
	conf := serviceConfig(serviceName)

	// 3. Configure Security (TLS vs Insecure)
	if conf.CertFile != "" {
		conf.CertFile = filepath.Join(utils.GetCertDir(), conf.CertFile)
		creds, err := credentials.NewClientTLSFromFile(conf.CertFile, serviceName)
		if err != nil {
			return nil, err
		}
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 4. Configure Keepalive parameters to maintain healthy long-lived connections.
	options = append(options, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                conf.Options.Keepalive.Time.Duration,
		Timeout:             conf.Options.Keepalive.Timeout.Duration,
		PermitWithoutStream: conf.Options.Keepalive.PermitWithoutStream,
	}))

	// 5. Setup Interceptor (Middleware) pipeline.
	// This wraps the framework's generic interceptor into gRPC's specific UnaryClientInterceptor format.
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

	// 6. Create the underlying gRPC client.
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
