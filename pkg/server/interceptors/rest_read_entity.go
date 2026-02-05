package interceptors

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"google.golang.org/protobuf/proto"
)

const (
	// RestReadEntityInterceptorName is the unique identifier for this interceptor.
	RestReadEntityInterceptorName = "restReadEntity"
)

func init() {
	// Register the interceptor specifically for the REST protocol.
	// This ensures that gRPC calls (which handle serialization natively) bypass this logic.
	server.AddInterceptor(RestReadEntityInterceptorName, NewReadEntityInterceptor, rest.Protocol)
}

// ReadEntity manages the logic for unmarshaling request data into Protobuf messages.
type ReadEntity struct {
	cm   sync.RWMutex
	conf ReadEntityConfig
}

// ReadEntityConfig defines which methods should bypass automatic parameter parsing.
type ReadEntityConfig struct {
	// SkipMethods is a list of full method names (e.g., /api.v1.User/Upload) to ignore.
	SkipMethods   utils.JSONStrings   `json:"skipMethods"`
	skipMethodMap map[string]struct{} // Map for O(1) lookup.
}

// Name returns the interceptor name.
func (r *ReadEntity) Name() string {
	return RestReadEntityInterceptorName
}

// NewReadEntityInterceptor initializes the interceptor and sets up dynamic configuration watching.
func NewReadEntityInterceptor() (server.ServerInterceptor, error) {
	readEntity := &ReadEntity{}
	if err := readEntity.loadAndWatch(); err != nil {
		return nil, err
	}
	return readEntity, nil
}

// Interceptor returns the actual middleware function.
func (r *ReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		logger.L(ctx).Debug("start server interceptor", "interceptor", r.Name(), "full_method", info.FullMethod, "protocol", info.Protocol)
		// 1. Verify that we are within a REST context and the method isn't skipped.
		rtx, ok := ctx.(*rest.Context)
		if ok && !r.isSkipped(info.FullMethod) {
			// 2. Automagically parse JSON/Form/Query data into the 'req' proto.Message.
			// This populates the request object that the business handler expects.
			if err := rtx.ReadEntity(req.(proto.Message)); err != nil {
				return nil, err
			}
		}

		// 3. Pass the populated request object to the next handler in the chain.
		return handler(ctx, req)
	}
}

// isSkipped checks if a specific method is exempted from automatic entity reading.
func (r *ReadEntity) isSkipped(method string) bool {
	r.cm.RLock()
	defer r.cm.RUnlock()
	_, ok := r.conf.skipMethodMap[method]
	return ok
}

// loadAndWatch handles the initial configuration and subscribes to updates from the config center.
func (r *ReadEntity) loadAndWatch() error {
	if err := r.load(); err != nil {
		return err
	}
	config.AddPrefixListener("asjard.interceptors.server.restReadEntity", r.watch)
	return nil
}

// load fetches the configuration and rebuilds the lookup map.
func (r *ReadEntity) load() error {
	conf := ReadEntityConfig{
		skipMethodMap: map[string]struct{}{},
	}
	if err := config.GetWithUnmarshal("asjard.interceptors.server.restReadEntity", &conf); err != nil {
		return err
	}
	for _, item := range conf.SkipMethods {
		conf.skipMethodMap[item] = struct{}{}
	}
	r.cm.Lock()
	r.conf = conf
	r.cm.Unlock()
	return nil
}

// watch is the callback for real-time configuration changes.
func (r *ReadEntity) watch(event *config.Event) {
	if err := r.load(); err != nil {
		logger.Error("read entity load config fail", "err", err)
	}
}
