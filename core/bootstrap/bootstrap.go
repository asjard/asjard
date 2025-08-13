/*
Package bootstrap 服务初始化后启动之前执行的一些初始化任务，加载一些内建功能
*/
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

// Initiator Initialization methods that need to be implemented
type Initiator interface {
	Start() error
	Stop()
}

var (
	bootstrapHandlers []Initiator
	bootstrapedMap    = make(map[Initiator]struct{})

	initiatorHandlers []Initiator
	initiatorMap      = make(map[Initiator]struct{})
)

// AddBootstrap adds the startup method
// Executed after initialization and before the service starts
func AddBootstrap(handler Initiator) {
	if _, ok := bootstrapedMap[handler]; !ok {
		bootstrapHandlers = append(bootstrapHandlers, handler)
		bootstrapedMap[handler] = struct{}{}
	}
}

// AddBootstraps Batch add startup method
func AddBootstraps(handlers ...Initiator) {
	for _, handler := range handlers {
		AddBootstrap(handler)
	}
}

// AddInitator adds initialization methods
// Loads into the env file environment variable and executes
func AddInitiator(handler Initiator) {
	if _, ok := initiatorMap[handler]; !ok {
		initiatorHandlers = append(initiatorHandlers, handler)
		initiatorMap[handler] = struct{}{}
	}
}

// AddInitiators Batch add initialization method
func AddInitiators(handlers ...Initiator) {
	for _, handler := range handlers {
		AddInitiator(handler)
	}
}

// Init run all initialization methods.
func Init() error {
	for _, handler := range initiatorHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Bootstrap run all startup methods.
func Bootstrap() error {
	for _, handler := range bootstrapHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown stop all initializaiton and startup methods.
func Shutdown() {
	for idx := len(bootstrapHandlers) - 1; idx >= 0; idx-- {
		bootstrapHandlers[idx].Stop()
	}
	for idx := len(initiatorHandlers) - 1; idx >= 0; idx-- {
		initiatorHandlers[idx].Stop()
	}
}
