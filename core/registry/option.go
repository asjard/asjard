package registry

import (
	"sync"

	"github.com/asjard/asjard/core/runtime"
)

// Options holds the criteria used to filter service instances from the cache.
// It also contains data for setting up real-time watches on those instances.
type Options struct {
	// pickFuncs is a slice of predicates. An instance must satisfy ALL of them to be selected.
	pickFuncs []PickFunc
	// watch is the callback executed when an instance matching these criteria changes.
	watch func(*Event)
	// watchName is a unique identifier for this specific listener.
	watchName string
}

// PickFunc is a predicate function. Returns true if the instance meets the requirement.
type PickFunc func(instance *Instance) bool

// Option defines the function signature for modifying the Options struct.
type Option func(opts *Options)

var (
	// global singleton for default filtering behavior.
	defaultOptions     *Options
	defaultOptionsOnce sync.Once
)

// NewOptions creates a new Options struct by applying all provided functional options.
func NewOptions(opts []Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// DefaultOptions provides a "Safe Default" filter.
// By default, it only selects services that share the same App, Region, and Environment
// as the current running process to ensure traffic stays within logical boundaries.
func DefaultOptions() *Options {
	defaultOptionsOnce.Do(func() {
		app := runtime.GetAPP()
		defaultOptions = &Options{}
		opts := []Option{
			WithApp(app.App),
			WithRegion(app.Region),
			WithEnvironment(app.Environment),
		}
		for _, opt := range opts {
			opt(defaultOptions)
		}
	})
	return defaultOptions
}

// --- Filter Functions (Predicate Builders) ---

// WithApp filters instances by Application name.
func WithApp(app string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.App == app
		})
	}
}

// WithPickFunc allows users to inject arbitrary custom filtering logic.
func WithPickFunc(pickFuns []PickFunc) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, pickFuns...)
	}
}

// WithRegion filters instances by physical region (e.g., "us-east-1").
func WithRegion(region string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Region == region
		})
	}
}

// WithEnvironment filters instances by lifecycle stage (e.g., "prod", "dev").
func WithEnvironment(environment string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Environment == environment
		})
	}
}

// WithServiceName filters instances by their specific service name (e.g., "order-api").
func WithServiceName(serviceName string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Instance.Name == serviceName
		})
	}
}

// WithInstanceID filters for one specific unique instance.
func WithInstanceID(instanceID string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Instance.ID == instanceID
		})
	}
}

// WithRegistryName filters based on which discovery source found the service (e.g., "etcd").
func WithRegistryName(registryName string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.DiscoverName == registryName
		})
	}
}

// WithProtocol filters for instances that support a specific protocol (e.g., "grpc", "rest").
func WithProtocol(protocol string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			for p := range instance.Service.Endpoints {
				if p == protocol {
					return true
				}
			}
			return false
		})
	}
}

// WithVersion filters instances by their semantic version (e.g., "v1.2.0").
func WithVersion(version string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			return instance.Service.Instance.Version == version
		})
	}
}

// WithMetadata filters instances based on custom key-value pairs in their metadata.
// It ensures that ALL key-values in the provided map match the instance's metadata.
func WithMetadata(metadata map[string]string) func(opts *Options) {
	return func(opts *Options) {
		opts.pickFuncs = append(opts.pickFuncs, func(instance *Instance) bool {
			for wantKey, wantValue := range metadata {
				isOk := false
				for key, value := range instance.Service.Instance.MetaData {
					if wantKey == key && wantValue == value {
						isOk = true
						break
					}
				}
				if !isOk {
					return false
				}
			}
			return true
		})
	}
}

// WithWatch enables the observer pattern. When the service cache changes,
// the callback will be executed if the instance matches the current filters.
func WithWatch(watchName string, callback func(*Event)) func(opts *Options) {
	return func(opts *Options) {
		opts.watch = callback
		opts.watchName = watchName
	}
}

// --- Helpers ---

func okPickFunc() PickFunc {
	return func(instance *Instance) bool {
		return true
	}
}

func (opts *Options) getPickFuncs() []PickFunc {
	return opts.pickFuncs
}
