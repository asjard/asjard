/*
Package pprof provides a built-in server implementation for Go's performance profiling tools.
By enabling this server, developers can analyze the application's runtime behavior
using 'go tool pprof' or by visiting the /debug/pprof/ endpoint.
*/
package pprof

import (
	"errors"
	"fmt"
	"net/http"

	// Import net/http/pprof for its side effects.
	// This automatically registers pprof handlers to http.DefaultServeMux.
	_ "net/http/pprof"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

const (
	// Protocol defines the unique identifier for the pprof server.
	Protocol = "pprof"
)

// PprofServer manages the lifecycle of the profiling HTTP server.
type PprofServer struct {
	conf server.Config
}

// Ensure PprofServer satisfies the core server.Server interface.
var _ server.Server = &PprofServer{}

func init() {
	// Register the pprof server creator with the framework's server registry.
	server.AddServer(Protocol, New)
}

// MustNew initializes a PprofServer instance with the provided configuration.
func MustNew(conf server.Config, options *server.ServerOptions) (server.Server, error) {
	return &PprofServer{
		conf: conf,
	}, nil
}

// New creates a new PprofServer instance by loading configuration from the framework.
func New(options *server.ServerOptions) (server.Server, error) {
	conf := server.Config{}
	// Retrieve pprof-specific configuration using the standard prefix.
	if err := config.GetWithUnmarshal(constant.ConfigServerPporfPrefix, &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// AddHandler is a no-op for PprofServer as handlers are registered automatically
// via the net/http/pprof package import.
func (s *PprofServer) AddHandler(_ any) error {
	return nil
}

// Start launches the pprof HTTP server in a separate goroutine.
func (s *PprofServer) Start(startErr chan error) error {
	// Ensure a listening address is configured.
	if s.conf.Addresses.Listen == "" {
		return errors.New("config servers.pprof.addresses.listen not found")
	}

	go func() {
		// ListenAndServe uses http.DefaultServeMux where pprof handlers are registered.
		if err := http.ListenAndServe(s.conf.Addresses.Listen, nil); err != nil {
			startErr <- fmt.Errorf("start pprof with address %s fail %s", s.conf.Addresses.Listen, err.Error())
		}
	}()

	logger.Debug("start pprof server", "address", s.conf.Addresses.Listen)
	return nil
}

// Stop terminates the pprof server.
// Note: http.ListenAndServe doesn't provide a simple Stop; in production,
// this is usually handled by the process exiting.
func (s *PprofServer) Stop() {}

// Protocol returns the protocol name "pprof".
func (s *PprofServer) Protocol() string {
	return Protocol
}

// Enabled checks if the pprof server should be started based on configuration.
func (s *PprofServer) Enabled() bool {
	return s.conf.Enabled
}

// ListenAddresses returns the network address configuration for the pprof server.
func (s *PprofServer) ListenAddresses() server.AddressConfig {
	return s.conf.Addresses
}
