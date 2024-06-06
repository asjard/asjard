package grpc

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

const (
	// Protocol 协议名称
	Protocol = "grpc"
)

// GrpcServer .
type GrpcServer struct {
	addresses   map[string]string
	enabled     bool
	server      *grpc.Server
	interceptor server.UnaryServerInterceptor
}

// Handler .
type Handler interface {
	GrpcServiceDesc() *grpc.ServiceDesc
}

var _ server.Server = &GrpcServer{}

func init() {
	server.AddServer(Protocol, New)
}

// New .
func New(interceptor server.UnaryServerInterceptor) (server.Server, error) {
	addressesMap := make(map[string]string)
	if err := config.GetWithUnmarshal("servers.grpc.addresses", &addressesMap); err != nil {
		return nil, err
	}
	var opts []grpc.ServerOption
	certFile := config.GetString("servers.grpc.certFile", "")
	if certFile != "" {
		certFile = filepath.Join(utils.GetCertDir(), certFile)
	}
	keyFile := config.GetString("servers.grpc.keyFile", "")
	if keyFile != "" {
		keyFile = filepath.Join(utils.GetCertDir(), keyFile)
	}
	if certFile != "" && keyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: config.GetDuration("servers.grpc.options.MaxConnectionIdle", 5*time.Minute),
		Time:              config.GetDuration("servers.grpc.options.Time", 10*time.Second),
		Timeout:           config.GetDuration("servers.grpc.options.Timeout", time.Second),
	}))
	opts = append(opts, grpc.ChainUnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if interceptor != nil {
			return interceptor(ctx, req, &server.UnaryServerInfo{
				Server:     info.Server,
				FullMethod: info.FullMethod,
				Protocol:   Protocol,
			}, func(ctx context.Context, in any) (any, error) {
				return handler(ctx, in)
			})
		}
		return handler(ctx, req)
	}))
	return &GrpcServer{
		addresses: addressesMap,
		enabled:   config.GetBool("servers.grpc.enabled", false),
		server:    grpc.NewServer(opts...),
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

// WithChainUnaryInterceptor 设置拦截器
func (s *GrpcServer) WithChainUnaryInterceptor(interceptor server.UnaryServerInterceptor) {
	s.interceptor = interceptor
}

// Start .
func (s *GrpcServer) Start(startErr chan error) error {
	address, ok := s.addresses[constant.ServerListenAddressName]
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
	logger.Debug("start grpc server",
		"address", address)
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
func (s *GrpcServer) ListenAddresses() map[string]string {
	return s.addresses
}
