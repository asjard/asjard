package pprof

import (
	"errors"
	"fmt"
	"net/http"

	// 初始化pprof
	_ "net/http/pprof"

	"github.com/asjard/asjard/core/config"
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
func New(interceptor server.UnaryServerInterceptor) (server.Server, error) {
	server := &PprofServer{
		addresses: make(map[string]string),
		enabled:   config.GetBool("servers.pprof.enabled", false),
	}
	if err := config.GetWithUnmarshal("servers.pprof.addresses", &server.addresses); err != nil {
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
	address, ok := s.addresses["listen"]
	if !ok {
		return errors.New("config servers.pprof.addresses.listen not found")
	}
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			startErr <- fmt.Errorf("start pprof with adress %s fail %s", address, err.Error())
		}
	}()
	logger.Debugf("start pprof server on address: %s", address)
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
func (s *PprofServer) ListenAddresses() []*server.EndpointAddress {
	var addresses []*server.EndpointAddress
	for name, address := range s.addresses {
		addresses = append(addresses, &server.EndpointAddress{
			Name:    name,
			Address: address,
		})
	}
	return addresses
}
