package grpc

import (
	"github.com/asjard/asjard/core/server"
)

const (
	// Protocol 协议名称
	Protocol = "grpc"
)

// GrpcServer .
type GrpcServer struct {
	addresses []*server.EndpointAddress
}

var _ server.Server = &GrpcServer{}

func init() {
	server.AddServer(New)
}

// New .
func New() (server.Server, error) {
	return &GrpcServer{
		addresses: []*server.EndpointAddress{
			{Name: "grpc", Address: ":8080"},
		},
	}, nil
}

// AddHandler .
func (s *GrpcServer) AddHandler(handler interface{}) error {
	return nil
}

// Handle .
func (s *GrpcServer) Handle(req *server.Request) (*server.Response, error) {
	return nil, nil
}

// Start .
func (s *GrpcServer) Start() error {
	return nil
}

// Stop .
func (s *GrpcServer) Stop() {
}

// Protocol .
func (s *GrpcServer) Protocol() string {
	return Protocol
}

// ListenAddresses 监听地址列表
func (s *GrpcServer) ListenAddresses() []*server.EndpointAddress {
	return s.addresses
}
