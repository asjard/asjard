package pprof

import (
	"errors"
	"fmt"
	"net/http"

	// 初始化pprof
	_ "net/http/pprof"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

const (
	// Protocol 协议名称
	Protocol = "pprof"
)

// PprofServer .
type PprofServer struct {
	conf server.Config
}

var _ server.Server = &PprofServer{}

func init() {
	server.AddServer(Protocol, New)
}

func MustNew(conf server.Config, options *server.ServerOptions) (server.Server, error) {
	return &PprofServer{
		conf: conf,
	}, nil
}

// New .
func New(options *server.ServerOptions) (server.Server, error) {
	conf := server.Config{}
	if err := config.GetWithUnmarshal(constant.ConfigServerPporfPrefix, &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// AddHandler .
func (s *PprofServer) AddHandler(_ any) error {
	return nil
}

// Start .
func (s *PprofServer) Start(startErr chan error) error {
	if s.conf.Addresses.Listen == "" {
		return errors.New("config servers.pprof.addresses.listen not found")
	}
	go func() {
		if err := http.ListenAndServe(s.conf.Addresses.Listen, nil); err != nil {
			startErr <- fmt.Errorf("start pprof with address %s fail %s", s.conf.Addresses.Listen, err.Error())
		}
	}()
	logger.Debug("start pprof server",
		"address", s.conf.Addresses.Listen)
	return nil
}

// Stop .
func (s *PprofServer) Stop() {}

// Protocol .
func (s *PprofServer) Protocol() string {
	return Protocol
}

// Enabled .
func (s *PprofServer) Enabled() bool {
	return s.conf.Enabled
}

// ListenAddresses .
func (s *PprofServer) ListenAddresses() server.AddressConfig {
	return s.conf.Addresses
}
