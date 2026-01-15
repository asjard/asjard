/*
Package rest provides a built-in HTTP client implementation.
Note: The full client functionality is currently a work in progress (WIP).
It aims to provide a high-performance REST client that integrates with the
framework's service discovery and load-balancing abstractions.
*/
package rest

import (
	"context"

	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

// ClientConnInterface defines the contract for invoking HTTP requests.
// It abstracts the underlying transport to allow for interceptors and
// standardized request handling.
type ClientConnInterface interface {
	// Invoke executes an HTTP request.
	// method: HTTP verb (GET, POST, etc.)
	// path: The resource path on the target service.
	Invoke(ctx context.Context, method, path string)
}

// Client is a wrapper around the fasthttp.Client.
// It implements gRPC's resolver.ClientConn to allow HTTP services
// to be discovered using the same naming resolvers (like ETCD or Consul)
// used by gRPC clients.
type Client struct {
	*fasthttp.Client
}

// Verify that Client satisfies the gRPC resolver.ClientConn interface at compile time.
var _ resolver.ClientConn = &Client{}

// New creates and returns a new REST client instance.
func New() *Client {
	return &Client{}
}

// NewAddress is called by the resolver when the list of available
// service addresses changes. (WIP)
func (c *Client) NewAddress(addresses []resolver.Address) {}

// ParseServiceConfig parses service-level configurations such as
// load balancing policies or timeouts. (WIP)
func (c *Client) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	return &serviceconfig.ParseResult{}
}

// ReportError allows the resolver to notify the client of discovery failures. (WIP)
func (c *Client) ReportError(err error) {}

// UpdateState updates the client's internal connection state based
// on information from the resolver. (WIP)
func (c *Client) UpdateState(state resolver.State) error {
	return nil
}
