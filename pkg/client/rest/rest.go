/*
Package rest 内建http客户端实现,暂未实现
*/
package rest

import (
	"context"

	"github.com/valyala/fasthttp"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

// ClientConnInterface .
type ClientConnInterface interface {
	Invoke(ctx context.Context, method, path string)
}

// Client .
type Client struct {
	*fasthttp.Client
}

var _ resolver.ClientConn = &Client{}

// New .
func New() *Client {
	return &Client{}
}

// NewAddress .
func (c *Client) NewAddress(addresses []resolver.Address) {}

// ParseServiceConfig .
func (c *Client) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	return &serviceconfig.ParseResult{}
}

// ReportError .
func (c *Client) ReportError(err error) {}

// UpdateState .
func (c *Client) UpdateState(state resolver.State) error {
	return nil
}
