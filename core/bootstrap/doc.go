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
