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
	addresses map[string]string
	enabled   bool
}

var _ server.Server = &PprofServer{}

func init() {
	server.AddServer(Protocol, New)
}

// New .
func New(options *server.ServerOptions) (server.Server, error) {
	server := &PprofServer{
		addresses: make(map[string]string),
		enabled:   config.GetBool(fmt.Sprintf(constant.ConfigServerEnabled, Protocol), false),
	}
	if err := config.GetWithUnmarshal(fmt.Sprintf(constant.ConfigServerAddress, Protocol), &server.addresses); err != nil {
		return server, err
	}
	return server, nil
}

// AddHandler .
func (s *PprofServer) AddHandler(_ any) error {
	return nil
}

// Start .
func (s *PprofServer) Start(startErr chan error) error {
	address, ok := s.addresses[constant.ServerListenAddressName]
	if !ok {
		return errors.New("config servers.pprof.addresses.listen not found")
	}
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			startErr <- fmt.Errorf("start pprof with adress %s fail %s", address, err.Error())
		}
	}()
	logger.Debug("start pprof server",
		"address", address)
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
	return s.enabled
}

// ListenAddresses .
func (s *PprofServer) ListenAddresses() map[string]string {
	return s.addresses
}
