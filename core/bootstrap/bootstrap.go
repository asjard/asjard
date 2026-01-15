/*
Package bootstrap manages the lifecycle orchestration and component coordination.

It abstracts the startup (Start) and cleanup (Stop) behaviors of components through
the Initiator interface and establishes two core execution phases:

 1. Initiator Phase: Base environment initialization. Typically used for loading
    environment variables, memory-based configurations, logging systems, and other
    low-level infrastructure.
 2. Bootstrap Phase: Functional component activation. Typically used for enabling
    security modules, interceptors, handlers, pprof, and service discovery.

Core Design Principles:
  - Plug-and-Play: Enables decoupled registration via init() functions using
    AddInitiator or AddBootstrap through side-effect imports.
  - Sequential Startup: Executes Start methods in the order of registration to
    ensure base dependencies are ready first.
  - Reverse Shutdown: The Shutdown function follows the LIFO (Last-In-First-Out)
    principle, stopping components in reverse order to ensure safe resource release.
  - Idempotency: Maintains internal state maps to prevent redundant registration
    of the same component.
*/
// Example:
//
//	type MyService struct {}
//	func (s *MyService) Start() error { return nil }
//	func (s *MyService) Stop() {}
//
//	func init() {
//	    bootstrap.AddBootstrap(&MyService{})
//	}
package bootstrap

import (
	// init security component
	_ "github.com/asjard/asjard/pkg/security"
	// init server interceptors
	_ "github.com/asjard/asjard/pkg/server/interceptors"
	// init server handlers
	_ "github.com/asjard/asjard/pkg/server/handlers"
	// init client interceptors
	_ "github.com/asjard/asjard/pkg/client/interceptors"
	// init pprof
	_ "github.com/asjard/asjard/pkg/server/pprof"
	// init mem configuration source
	_ "github.com/asjard/asjard/pkg/config/mem"
	// init env configuration source
	_ "github.com/asjard/asjard/pkg/config/env"
)

// Initiator defines the lifecycle contract for components within the framework.
// Any module that requires setup at startup or cleanup at shutdown should
// implement this interface.
type Initiator interface {
	// Start executes the initialization or startup logic for the component.
	// If Start returns an error, the bootstrapping process will typically be aborted.
	Start() error
	// Stop handles the graceful teardown of the component.
	// It is called during the shutdown phase to release resources,
	// close connections, or stop background goroutines.
	Stop()
}

var (
	// bootstrapHandlers stores tasks for the functional component activation phase.
	bootstrapHandlers []Initiator
	// bootstrapedMap ensures idempotency for bootstrap tasks.
	bootstrapedMap = make(map[Initiator]struct{})

	// initiatorHandlers stores tasks for the base environment initialization phase.
	initiatorHandlers []Initiator
	// initiatorMap ensures idempotency for initiator tasks.
	initiatorMap = make(map[Initiator]struct{})
)

// AddBootstrap registers a handler for the Bootstrap phase.
// These handlers are executed after basic initialization but before the main service starts.
func AddBootstrap(handler Initiator) {
	if _, ok := bootstrapedMap[handler]; !ok {
		bootstrapHandlers = append(bootstrapHandlers, handler)
		bootstrapedMap[handler] = struct{}{}
	}
}

// AddBootstraps registers multiple handlers for the Bootstrap phase in a single call.
func AddBootstraps(handlers ...Initiator) {
	for _, handler := range handlers {
		AddBootstrap(handler)
	}
}

// AddInitiator registers a handler for the Initiator phase.
// These handlers are used for low-level tasks like environment and config loading.
func AddInitiator(handler Initiator) {
	if _, ok := initiatorMap[handler]; !ok {
		initiatorHandlers = append(initiatorHandlers, handler)
		initiatorMap[handler] = struct{}{}
	}
}

// AddInitiators registers multiple handlers for the Initiator phase in a single call.
func AddInitiators(handlers ...Initiator) {
	for _, handler := range handlers {
		AddInitiator(handler)
	}
}

// Init executes all registered Initiator handlers sequentially.
// Returns the first error encountered, if any.
func Init() error {
	for _, handler := range initiatorHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Bootstrap executes all registered Bootstrap handlers sequentially.
// Typically called after successful execution of Init().
func Bootstrap() error {
	for _, handler := range bootstrapHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown gracefully stops all registered components in reverse order (LIFO).
// It first stops bootstrap components, then initialization components.
func Shutdown() {
	// Stop bootstrap handlers in reverse order
	for idx := len(bootstrapHandlers) - 1; idx >= 0; idx-- {
		bootstrapHandlers[idx].Stop()
	}
	// Stop initiator handlers in reverse order
	for idx := len(initiatorHandlers) - 1; idx >= 0; idx-- {
		initiatorHandlers[idx].Stop()
	}
}
