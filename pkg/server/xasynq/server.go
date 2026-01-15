package xasynq

import (
	"context"
	"fmt"
	"reflect"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/hibiken/asynq"
)

const (
	// Protocol is the identifier used to register this server type in the framework.
	Protocol = "asynq"
)

// AsynqServer defines the background task worker server.
// It manages an internal asynq.Server for task processing and a mux for routing.
type AsynqServer struct {
	srv     *asynq.Server         // The underlying Asynq worker engine.
	mux     *asynq.ServeMux       // The router that maps task types to handler functions.
	conf    Config                // Server-specific configuration (Redis, Concurrency, etc.).
	options *server.ServerOptions // Framework-level server options (Interceptors).
	app     runtime.APP           // Reference to the global application state.
}

var (
	// Ensure AsynqServer satisfies the core Server interface.
	_ server.Server = &AsynqServer{}
	// globalHandler holds the default hooks for retries, errors, and health checks.
	globalHandler GlobaltHandler = &defaultGlobalHandler{}
)

func init() {
	// Automatically register this server factory during application startup.
	server.AddServer(Protocol, New)
}

// New initializes the server by fetching configuration from the "asjard.servers.asynq" key.
func New(options *server.ServerOptions) (server.Server, error) {
	conf := defaultConfig()
	if err := config.GetWithUnmarshal("asjard.servers.asynq", &conf); err != nil {
		return nil, err
	}
	return MustNew(conf, options)
}

// MustNew builds the server instance, establishes Redis connections, and configures the worker.
func MustNew(conf Config, options *server.ServerOptions) (server.Server, error) {
	if !conf.Enabled {
		return &AsynqServer{}, nil
	}
	// Initialize the Redis connection pool for Asynq.
	redisConn, err := NewRedisConn(conf.Redis)
	if err != nil {
		return nil, err
	}
	return &AsynqServer{
		conf:    conf,
		options: options,
		app:     runtime.GetAPP(),
		srv: asynq.NewServer(redisConn, asynq.Config{
			Concurrency:     conf.Options.Concurrency,
			BaseContext:     globalHandler.BaseContext(),
			RetryDelayFunc:  globalHandler.RetryDelayFunc(),
			IsFailure:       globalHandler.IsFailure(),
			ErrorHandler:    globalHandler.ErrorHandler(),
			Logger:          &asynqLogger{}, // Injects the adapter to unify logs.
			HealthCheckFunc: globalHandler.HealthCheckFunc(),
			GroupAggregator: globalHandler.GroupAggregator(),
		}),
		mux: asynq.NewServeMux(),
	}, nil
}

// AddHandler validates that the provided service implements the Asynq Handler interface.
func (s *AsynqServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invalid handler, %T must implement *asynq.Handler", handler)
	}
	return s.addHandler(h)
}

// Start kicks off the Asynq worker loop in a background goroutine.
func (s *AsynqServer) Start(startErr chan error) error {
	go func() {
		// Run blocks until the server is stopped or a fatal error occurs.
		if err := s.srv.Run(s.mux); err != nil {
			startErr <- fmt.Errorf("start asynq fail %v", err)
		}
	}()
	return nil
}

// Stop initiates a graceful shutdown, allowing active tasks to complete.
func (s *AsynqServer) Stop() {
	s.srv.Stop()     // Stop fetching new tasks.
	s.srv.Shutdown() // Wait for active tasks to finish.
}

// Protocol returns "asynq".
func (s *AsynqServer) Protocol() string {
	return Protocol
}

// ListenAddresses is empty for Asynq as it doesn't open a network port (it pulls from Redis).
func (s *AsynqServer) ListenAddresses() server.AddressConfig {
	return server.AddressConfig{}
}

// Enabled returns true if the server is configured to run.
func (s *AsynqServer) Enabled() bool {
	return s.conf.Enabled
}

// addHandler registers all methods defined in the service descriptor to the router.
func (s *AsynqServer) addHandler(handler Handler) error {
	desc := handler.AsynqServiceDesc()
	if desc == nil {
		return nil
	}
	// Use reflection to ensure the handler actually satisfies the interface defined in the descriptor.
	ht := reflect.TypeOf(desc.HandlerType).Elem()
	st := reflect.TypeOf(handler)
	if !st.Implements(ht) {
		return fmt.Errorf("found the handler of type %v that does not satisfy %v", st, ht)
	}

	for _, method := range desc.Methods {
		if method.Pattern != "" && method.Handler != nil {
			s.addRouterHandler(method.Pattern, handler, method.Handler)
		}
	}
	return nil
}

// addRouterHandler wraps the business logic with Asjard's context and interceptor chain.
func (s *AsynqServer) addRouterHandler(fullMethodName string, svc Handler, handler handlerFunc) {
	s.mux.HandleFunc(Pattern(fullMethodName),
		func(ctx context.Context, task *asynq.Task) error {
			// Converts the standard context and task into an Asjard-compatible xasynq.Context.
			// This allows interceptors (logging, tracing) to run on background tasks.
			if _, err := handler(&Context{Context: ctx, task: task}, svc, s.options.Interceptor); err != nil {
				return err
			}
			return nil
		})
}
