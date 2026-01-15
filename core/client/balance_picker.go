package client

import (
	"sync"

	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
)

const (
	// HeaderLoadBalancer records which load balancing strategy was used for this request.
	// Useful for observability, logging, and debugging across services.
	HeaderLoadBalancer = "x-request-balancer"
)

// Picker defines the interface that all custom load balancers must implement.
// It is a higher-level abstraction over gRPC's native balancer.Picker.
type Picker interface {
	// Pick selects a backend connection based on the pick information.
	Pick(info balancer.PickInfo) (*PickResult, error)

	// Name returns the name of the load balancing strategy (used in logs and headers).
	Name() string
}

// PickResult extends the selection result with region/awareness information,
// enabling locality-aware routing decisions.
type PickResult struct {
	SubConn       *SubConn
	RequestRegion string // Desired region for routing (may come from context or default)
	RequestAz     string // Desired availability zone (may come from context or default)
}

// WrapPicker wraps every custom Picker to provide:
// - concurrency safety via mutex
// - consistent injection of request context metadata (app, source, region, etc.)
// - conversion to gRPC's native balancer.PickResult
type WrapPicker struct {
	app runtime.APP

	mu     sync.Mutex // Protects picker field (in case of dynamic updates)
	picker Picker
}

// PickerBase provides reusable helper methods for most custom pickers.
type PickerBase struct {
	app runtime.APP
}

// NewPickerBase creates a new PickerBase instance with current app runtime info.
func NewPickerBase(app runtime.APP) *PickerBase {
	return &PickerBase{app: app}
}

// CanReachable determines if the given SubConn is reachable from the current node.
//
// Rules:
// - Same region → prefer internal listen address (low latency, no public routing)
// - Different region → only allow advertised address (public/controlled entry point)
func (p PickerBase) CanReachable(sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if !ok {
		return false
	}

	// Same region: use listen address (internal network)
	if sc.Address.Attributes.Value(ListenAddressKey{}) != nil &&
		instance.Service.Region == p.app.Region {
		return true
	}

	// Cross-region: only allow advertised address (public or dedicated link)
	if sc.Address.Attributes.Value(AdvertiseAddressKey{}) != nil &&
		instance.Service.Region != p.app.Region {
		return true
	}

	return false
}

// Shareable checks whether the instance allows traffic from other regions/clusters.
// Used to enforce sharing policies in multi-tenant or staged rollout scenarios.
func (p PickerBase) Shareable(sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if !ok {
		return false
	}
	return instance.Service.Instance.Shareable
}

// NewPicker is the factory function used by custom balancer builders
// to create a wrapped Picker with consistent behavior.
func NewPicker(newBalancerPicker NewBalancerPicker, scs map[balancer.SubConn]base.SubConnInfo) balancer.Picker {
	return &WrapPicker{
		picker: newBalancerPicker(scs),
		app:    runtime.GetAPP(),
	}
}

// Pick selects a connection and enriches the result with request metadata.
//
// It:
// 1. Delegates to the underlying strategy-specific picker
// 2. Fills in default region/AZ if not specified
// 3. Injects observability headers for full request context
func (p *WrapPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	result, err := p.picker.Pick(info)
	p.mu.Unlock()
	if err != nil {
		return balancer.PickResult{}, err
	}

	// Extract destination service name for observability
	destServiceName := ""
	if instance, ok := result.SubConn.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance); ok {
		destServiceName = instance.Service.Instance.Name
	}

	// Use request-specified region, fall back to local
	requestRegion := result.RequestRegion
	if requestRegion == "" {
		requestRegion = p.app.Region
	}

	// Use request-specified AZ, fall back to local
	requestAz := result.RequestAz
	if requestAz == "" {
		requestAz = p.app.AZ
	}

	return balancer.PickResult{
		SubConn: result.SubConn.Conn,

		// Done callback is currently a no-op.
		// Consider implementing it later for stats, health tracking, or dynamic weighting.
		Done: func(info balancer.DoneInfo) {},

		// Attach rich context to outgoing metadata for:
		// - distributed tracing
		// - logging & metrics
		// - server-side routing decisions
		Metadata: metadata.New(map[string]string{
			HeaderRequestApp:    p.app.App,
			HeaderRequestSource: p.app.Instance.Name,
			HeaderRequestDest:   destServiceName,
			HeaderRequestRegion: requestRegion,
			HeaderRequestAz:     requestAz,
			HeaderLoadBalancer:  p.picker.Name(),
		}),
	}, nil
}
