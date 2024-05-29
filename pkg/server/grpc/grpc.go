package grpc

import (
	"errors"
	"net"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	// Protocol 协议名称
	Protocol = "grpc"
)

// GrpcServer .
type GrpcServer struct {
	addresses map[string]string
	enabled   bool
	server    *grpc.Server
}

// Handler .
type Handler interface {
	GrpcServiceDesc() *grpc.ServiceDesc
}

var _ server.Server = &GrpcServer{}

func init() {
	server.AddServer(New)
}

// New .
func New() (server.Server, error) {
	addressesMap := make(map[string]string)
	if err := config.GetWithUnmarshal("servers.grpc.addresses", &addressesMap); err != nil {
		return nil, err
	}
	return &GrpcServer{
		addresses: addressesMap,
		server: grpc.NewServer(grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: config.GetDuration("servers.grpc.options.MaxConnectionIdle", 5*time.Minute),
			Time:              config.GetDuration("servers.grpc.options.Time", 10*time.Second),
			Timeout:           config.GetDuration("servers.grpc.options.Timeout", time.Second),
		})),
		enabled: config.GetBool("servers.grpc.enabled", false),
	}, nil
}

// AddHandler .
func (s *GrpcServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return errors.New("invalid handler, must implement *grpc.ServiceDesc")
	}
	s.server.RegisterService(h.GrpcServiceDesc(), handler)
	return nil
}

// Start .
func (s *GrpcServer) Start(startErr chan error) error {
	address, ok := s.addresses["listen"]
	if !ok {
		return errors.New("config servers.grpc.addresses.listen not found")
	}
	listen, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	go func() {
		if err := s.server.Serve(listen); err != nil {
			startErr <- err
		}
	}()
	logger.Debugf("start grpc server on address: %s", address)
	return nil
}

// Stop .
func (s *GrpcServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Protocol .
func (s *GrpcServer) Protocol() string {
	return Protocol
}

// Enabled .
func (s *GrpcServer) Enabled() bool {
	return s.enabled
}

// ListenAddresses 监听地址列表
func (s *GrpcServer) ListenAddresses() []*server.EndpointAddress {
	var addresses []*server.EndpointAddress
	for name, address := range s.addresses {
		addresses = append(addresses, &server.EndpointAddress{
			Name:    name,
			Address: address,
		})
	}
	return addresses
}
