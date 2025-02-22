package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"path/filepath"

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
	server *grpc.Server
	conf   Config
}

// ServiceDesc gpc.ServiceDesc别名
type ServiceDesc = grpc.ServiceDesc

// Handler .
type Handler interface {
	GrpcServiceDesc() *ServiceDesc
}

var _ server.Server = &GrpcServer{}

func init() {
	server.AddServer(Protocol, New)
}

func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	var opts []grpc.ServerOption
	if conf.CertFile != "" && conf.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(conf.CertFile, conf.KeyFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Creds(creds))
	}
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionIdle: conf.Options.KeepaliveParams.MaxConnectionIdle.Duration,
		Time:              conf.Options.KeepaliveParams.Time.Duration,
		Timeout:           conf.Options.KeepaliveParams.Timeout.Duration,
	}))
	opts = append(opts, grpc.ChainUnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if options.Interceptor != nil {
			return options.Interceptor(ctx, req, &server.UnaryServerInfo{
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
		server: grpc.NewServer(opts...),
		conf:   conf,
	}, nil
}

// New .
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	if err := config.GetWithUnmarshal(constant.ConfigServerGrpcPrefix, &conf); err != nil {
		return nil, err
	}
	if conf.KeyFile != "" {
		conf.KeyFile = filepath.Join(utils.GetCertDir(), conf.KeyFile)
	}
	if conf.CertFile != "" {
		conf.CertFile = filepath.Join(utils.GetCertDir(), conf.CertFile)
	}
	return MustNew(conf, options)

}

// AddHandler .
func (s *GrpcServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invalid handler %T, must implement *grpc.ServiceDesc", handler)
	}
	s.server.RegisterService(h.GrpcServiceDesc(), handler)
	return nil
}

// Start .
func (s *GrpcServer) Start(startErr chan error) error {
	if s.conf.Addresses.Listen == "" {
		return errors.New("config servers.grpc.addresses.listen not found")
	}
	listen, err := net.Listen("tcp", s.conf.Addresses.Listen)
	if err != nil {
		return err
	}
	go func() {
		if err := s.server.Serve(listen); err != nil {
			startErr <- err
		}
	}()
	logger.Debug("start grpc server",
		"address", s.conf.Addresses.Listen)
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
	return s.conf.Enabled
}

// ListenAddresses 监听地址列表
func (s *GrpcServer) ListenAddresses() server.AddressConfig {
	return s.conf.Addresses
}
