package asjard

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/config/sources/file"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/metrics"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/server/handlers"
	"github.com/asjard/asjard/core/trace"
	"github.com/asjard/asjard/utils"
)

const (
	website = "https://github.com/asjard/asjard"
	// The banner displayed on startup containing environment and instance metadata.
	banner = `
                                       App:      %s
                                       Env:      %s
    _   ___    _  _   ___ ___          Region:   %s
   /_\ / __|_ | |/_\ | _ \   \         Az:       %s
  / _ \\__ \ || / _ \|   / |) |
 /_/ \_\___/\__/_/ \_\_|_\___/ %s
                                       ID:       %s
                                       Name:     %s
                                       Version:  %s
                                       Servers:  %s
                                       ConfDir:  %s
 `
)

// Asjard manages the registered servers and their respective handlers.
// It acts as the central controller for the framework's startup and shutdown phases.
type Asjard struct {
	// servers is the list of protocol servers (e.g., gRPC, HTTP) to be managed.
	servers []server.Server
	// handlers maps protocol names to their business logic handlers.
	handlers map[string][]any
	hm       sync.RWMutex
	// startErr captures asynchronous errors during the server startup process.
	startErr chan error
	// startedServers tracks the strings of successfully bound addresses.
	startedServers []string
	// inited ensures the framework initialization logic runs only once.
	inited atomic.Bool
}

// New creates a new instance of the Asjard orchestrator.
func New() *Asjard {
	return &Asjard{
		handlers: make(map[string][]any),
		startErr: make(chan error),
	}
}

// AddHandler associates a business logic handler with one or more protocols.
// If the handler implements bootstrap.Initiator, it is automatically added to the bootstrap sequence.
func (asd *Asjard) AddHandler(handler any, protocols ...string) error {
	asd.hm.Lock()
	defer asd.hm.Unlock()
	for _, protocol := range protocols {
		if _, ok := asd.handlers[protocol]; ok {
			asd.handlers[protocol] = append(asd.handlers[protocol], handler)
		} else {
			asd.handlers[protocol] = []any{handler}
		}
		// Register as a bootstrap initiator if supported.
		if bootstrapHandler, ok := handler.(bootstrap.Initiator); ok {
			bootstrap.AddBootstrap(bootstrapHandler)
		}
	}
	return nil
}

// AddHandlers is a helper function to add multiple handlers to a single protocol.
func (asd *Asjard) AddHandlers(protocol string, handlers ...any) error {
	for _, handler := range handlers {
		if err := asd.AddHandler(handler, protocol); err != nil {
			return err
		}
	}
	return nil
}

// Start triggers the system startup sequence:
// 1. Initialization 2. Discovery 3. Server start 4. Registry registration.
// It also blocks until a termination signal is received for graceful shutdown.
func (asd *Asjard) Start() error {
	logger.Info("System Starting...")
	if err := asd.Init(); err != nil {
		return err
	}

	// Discover existing cluster services.
	if err := registry.Discover(); err != nil {
		return err
	}

	// Start protocol listeners (HTTP/gRPC, etc.).
	if err := asd.startServers(); err != nil {
		return err
	}

	// Announce this instance to the service registry (Consul, Etcd, etc.).
	if err := registry.Registe(); err != nil {
		return err
	}

	logger.Info("System Started")
	if !config.GetBool(constant.ConfigLoggerBannerDisable, false) {
		asd.printBanner()
	}

	// Wait for OS signals or startup errors.
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case s := <-quit:
		logger.Info("system get os signal start exiting...",
			"signal", s.String())
	case err := <-asd.startErr:
		logger.Error("start error:",
			"error", err)
	}

	// Standard delay to allow service meshes or load balancers to detect the offline state.
	time.Sleep(5 * time.Second)
	close(runtime.Exit)

	// Execute shutdown sequence.
	asd.stop()
	return nil
}

// Exit returns a channel that is closed when the system begins its shutdown process.
func (asd *Asjard) Exit() <-chan struct{} {
	return runtime.Exit
}

// Init handles the orderly initialization of core framework components.
// Order: File Config -> Bootstrap Inits -> Remote Config -> Metrics/Tracing -> Clients -> Registry.
func (asd *Asjard) Init() error {
	if asd.inited.Load() {
		return nil
	}
	defer asd.inited.Store(true)

	// Load initial local file configuration (highest priority).
	if err := config.Load(file.Priority); err != nil {
		return err
	}

	// Initialize basic bootstrap components.
	if err := bootstrap.Init(); err != nil {
		return err
	}

	// Load remaining configuration sources.
	if err := config.Load(-1); err != nil {
		return err
	}

	// Initialize observability (metrics and tracing).
	if err := metrics.Init(); err != nil {
		return err
	}
	if err := trace.Init(); err != nil {
		return err
	}

	// Initialize internal clients and service registry.
	if err := client.Init(); err != nil {
		return err
	}
	if err := registry.Init(); err != nil {
		return err
	}

	// Run standard bootstrap tasks.
	if err := bootstrap.Bootstrap(); err != nil {
		return err
	}

	// Prepare the protocol servers.
	servers, err := server.Init()
	if err != nil {
		return err
	}
	asd.servers = servers
	return nil
}

// startServers iterates through all enabled servers, binds handlers, and begins listening.
func (asd *Asjard) startServers() error {
	logger.Debug("Start start servers")
	defer logger.Debug("start servers Done")
	svc := server.GetService()

	for _, sv := range asd.servers {
		if !sv.Enabled() {
			continue
		}
		logger.Debug("Start start server", "protocol", sv.Protocol())

		// Map user-defined handlers to the server.
		asd.hm.RLock()
		for _, handler := range asd.handlers[sv.Protocol()] {
			if err := sv.AddHandler(handler); err != nil {
				return fmt.Errorf("server '%s' add handler fail[%s]", sv.Protocol(), err.Error())
			}
		}
		asd.hm.RUnlock()

		// Attach framework-level default handlers (e.g., health checks).
		for _, handler := range handlers.GetServerDefaultHandlers(sv.Protocol()) {
			if err := sv.AddHandler(handler); err != nil {
				return fmt.Errorf("server %s add default handler fail[%s]", sv.Protocol(), err.Error())
			}
		}

		// Calculate listen and advertise addresses for discovery.
		listenAddresses := sv.ListenAddresses()
		if err := svc.AddEndpoint(sv.Protocol(), listenAddresses); err != nil {
			return fmt.Errorf("server '%s' add endpoint fail[%s]", sv.Protocol(), err.Error())
		}

		// Trigger the actual listener.
		if err := sv.Start(asd.startErr); err != nil {
			return fmt.Errorf("start server '%s' fail[%s]", sv.Protocol(), err.Error())
		}

		protocolPrefix := sv.Protocol() + "://"
		asd.startedServers = append(asd.startedServers,
			strings.TrimSuffix(strings.Join([]string{
				protocolPrefix + listenAddresses.Listen,
				protocolPrefix + listenAddresses.Advertise,
			}, ","), ","+protocolPrefix))
		logger.Debug("server started", "protocol", sv.Protocol())
	}
	return nil
}

// stop manages the graceful exit:
// 1. Deregister from discovery 2. Close servers 3. Disconnect config 4. Final cleanup.
func (asd *Asjard) stop() {
	logger.Debug("start remove instance from registry")
	// Step 1: Tell the registry this instance is gone so it stops receiving traffic.
	if err := registry.Unregiste(); err != nil {
		logger.Error("unregiste from registry fail", "error", err.Error())
	}

	// Step 2: Shut down all active servers.
	for _, server := range asd.servers {
		if server.Enabled() {
			logger.Debug("start stop server", "protocol", server.Protocol())
			server.Stop()
			logger.Debug("server stopped", "protocol", server.Protocol())
		}
	}

	time.Sleep(time.Second)
	// Finalize component shutdowns.
	config.Disconnect()
	bootstrap.Shutdown()
	logger.Info("system exited")
}

// printBanner displays the framework ASCII art and runtime metadata.
func (asd *Asjard) printBanner() {
	app := runtime.GetAPP()
	fmt.Printf(banner,
		app.App,
		app.Environment,
		app.Region,
		app.AZ,
		constant.FrameworkVersion,
		app.Instance.ID,
		app.Instance.Name,
		app.Instance.Version,
		strings.Join(asd.startedServers, ";"),
		utils.GetConfDir())
}
