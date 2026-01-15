/*
Package client manages client implementations for various protocols.
It provides unified connection management through configuration, dynamic
loading of interceptors, and selection of load balancing strategies.
*/
package client

import (
	"fmt"
	"net/url"

	"github.com/asjard/asjard/core/constant"
	"google.golang.org/grpc"
)

// ClientInterface defines the factory interface for creating protocol-specific connections.
type ClientInterface interface {
	// NewConn creates a new client connection.
	// Target format follows: asjard://protocol/serviceName
	NewConn(target string, options *ClientOptions) (ClientConnInterface, error)
}

// ClientConnInterface extends gRPC's ClientConnInterface to provide
// additional metadata about the connection.
type ClientConnInterface interface {
	grpc.ClientConnInterface
	// ServiceName returns the name of the remote service being called.
	ServiceName() string
	// Protocol returns the communication protocol (e.g., grpc, http).
	Protocol() string
	// Conn returns the underlying gRPC client connection.
	Conn() grpc.ClientConnInterface
}

// ConnOptions defines parameters for individual connection requests.
type ConnOptions struct {
	// InstanceID identifies a specific instance to connect to, bypassing load balancing.
	InstanceID string
	// RegistryName specifies which service discovery registry to use.
	RegistryName string
}

// ConnOption is a functional option pattern for configuring ConnOptions.
type ConnOption func(opts *ConnOptions)

// NewClientFunc is a factory function type that initializes a protocol-specific client.
type NewClientFunc func(*ClientOptions) ClientInterface

var (
	// newClients stores registered client constructors indexed by protocol.
	newClients = make(map[string]NewClientFunc)
	// clients stores initialized client instances.
	clients = make(map[string]ClientInterface)
)

// AddClient registers a new protocol client implementation to the global registry.
func AddClient(protocol string, newClient NewClientFunc) {
	newClients[protocol] = newClient
}

// Init initializes all registered protocol clients.
// It loads configurations, assembles interceptor chains, and builds resolvers/balancers.
func Init() error {
	for protocol, newClient := range newClients {
		conf := GetConfigWithProtocol(protocol)
		interceptor, err := getChainUnaryInterceptors(protocol, conf)
		if err != nil {
			return err
		}
		// Each protocol client is initialized with its own resolver and balancer builders.
		clients[protocol] = newClient(&ClientOptions{
			Resolver:    &ClientBuilder{},
			Balancer:    NewBalanceBuilder(conf.Loadbalance),
			Interceptor: interceptor,
		})
	}
	return nil
}

// Client represents a high-level handle to a remote service.
type Client struct {
	protocol   string
	serverName string
	conf       Config
}

// NewClient creates a new Client handle for a specific service and protocol.
func NewClient(protocol, serverName string) *Client {
	return &Client{
		protocol:   protocol,
		serverName: serverName,
	}
}

// Conn establishes or returns a connection to the target service.
// It dynamically applies service-specific interceptors and load balancing strategies.
func (c Client) Conn(ops ...ConnOption) (grpc.ClientConnInterface, error) {
	cc, ok := clients[c.protocol]
	if !ok {
		return nil, fmt.Errorf("protocol %s client not found", c.protocol)
	}

	// Fetch specific configuration for the target service to override defaults.
	conf := serviceConfig(c.protocol, c.serverName)
	interceptor, err := getChainUnaryInterceptors(c.protocol, conf)
	if err != nil {
		return nil, err
	}

	// Prepare options with service-specific balancer and interceptor chains.
	options := &ClientOptions{
		Balancer:    NewBalanceBuilder(conf.Loadbalance),
		Interceptor: interceptor,
	}

	// Construct the target URI: asjard://protocol/serviceName?instanceID=xxx
	target := fmt.Sprintf("%s://%s/%s?%s",
		constant.Framework, c.protocol, c.serverName, c.connOptions(ops...).queryString())

	return cc.NewConn(target, options)
}

// connOptions merges multiple functional options into a single ConnOptions struct.
func (c Client) connOptions(ops ...ConnOption) *ConnOptions {
	options := &ConnOptions{}
	for _, op := range ops {
		op(options)
	}
	return options
}

// queryString converts connection options into a URL-encoded query string for the resolver.
func (o ConnOptions) queryString() string {
	v := make(url.Values)
	v.Set("instanceID", o.InstanceID)
	// Note: RegistryName is handled by the resolver but can be added here if needed.
	return v.Encode()
}

// WithInstanceID sets a specific instance ID to target for the connection.
func WithInstanceID(instanceID string) func(opts *ConnOptions) {
	return func(opts *ConnOptions) {
		opts.InstanceID = instanceID
	}
}

// WithRegistryName specifies a custom registry for service discovery.
func WithRegistryName(registryName string) func(opts *ConnOptions) {
	return func(opts *ConnOptions) {
		opts.RegistryName = registryName
	}
}
