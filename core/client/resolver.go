package client

import (
	"fmt"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

// AddressAttrKey is an empty struct used as a key for storing registry.Instance
// metadata within RPC address attributes.
type AddressAttrKey struct{}

// ListenAddressKey is used as a marker key in address attributes to identify
// a local listening address.
type ListenAddressKey struct{}

// AdvertiseAddressKey is used as a marker key in address attributes to identify
// a publicly accessible broadcast (advertise) address.
type AdvertiseAddressKey struct{}

// ClientBuilder implements the RPC resolver.Builder interface.
// It parses the custom Asjard scheme to create a resolver that watches for service changes.
type ClientBuilder struct{}

var _ resolver.Builder = &ClientBuilder{}

// Build creates a new resolver for the given target.
// The target URL format is: asjard://[protocol]/[serviceName]?instanceID=[id]&registryName=[name]
func (c *ClientBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	query := target.URL.Query()
	cr := &clientResolver{
		cc:           cc,
		protocol:     target.URL.Host,
		serviceName:  target.Endpoint(),
		instanceID:   query.Get("instanceID"),
		registryName: query.Get("registryName"),
	}

	// Trigger initial resolution immediately upon building.
	cr.ResolveNow(resolver.ResolveNowOptions{})
	return cr, nil
}

// Scheme returns the naming scheme supported by this builder (default: "asjard").
func (*ClientBuilder) Scheme() string {
	return constant.Framework
}

// clientResolver implements the gRPC resolver.Resolver interface.
// It bridges the gRPC connection and the Asjard service discovery system.
type clientResolver struct {
	cc resolver.ClientConn
	// protocol defines the communication method (e.g., grpc, http).
	protocol string
	// serviceName is the name of the target service in the registry.
	serviceName string
	// instanceID optionally filters for a specific service instance.
	instanceID string
	// registryName optionally specifies which discovery backend to query.
	registryName string
}

// Close cleans up the resolver by removing the listener from the registry.
func (r *clientResolver) Close() {
	registry.RemoveListener(r.listenerName())
}

// ResolveNow fetches the latest service list from the discovery center.
// It configures registry options based on the current application runtime and target parameters.
func (r *clientResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	app := runtime.GetAPP()
	options := []registry.Option{
		registry.WithServiceName(r.serviceName),
		registry.WithProtocol(r.protocol),
		registry.WithEnvironment(app.Environment),
		registry.WithApp(app.App),
		// Watch enables real-time updates when instances join or leave.
		registry.WithWatch(r.listenerName(), r.watch),
	}
	if r.instanceID != "" {
		options = append(options, registry.WithInstanceID(r.instanceID))
	}
	if r.registryName != "" {
		options = append(options, registry.WithRegistryName(r.registryName))
	}
	instances := registry.PickServices(options...)
	r.update(instances)
}

// listenerName generates a unique identifier for the registry watch event.
func (r *clientResolver) listenerName() string {
	return fmt.Sprintf("%s_clientResolver_%s_%s",
		constant.Framework,
		r.protocol,
		r.serviceName)
}

// update transforms registry instances into gRPC-compatible resolver addresses
// and pushes them to the gRPC ClientConn.
func (r *clientResolver) update(instances []*registry.Instance) {
	addresses := []resolver.Address{}
	for _, instance := range instances {
		attr := attributes.New(AddressAttrKey{}, instance)
		endpoint, ok := instance.Service.GetEndpoint(r.protocol)
		if ok {
			// Process both Listen and Advertise addresses for the instance.
			if len(endpoint.Listen) != 0 {
				for _, addr := range endpoint.Listen {
					logger.Debug("client resolver add addr", "addr", addr)
					addresses = append(addresses, resolver.Address{
						Addr:       addr,
						Attributes: attr.WithValue(ListenAddressKey{}, true),
					})
				}
			}
			if len(endpoint.Advertise) != 0 {
				for _, addr := range endpoint.Advertise {
					logger.Debug("client resolver add addr", "addr", addr)
					addresses = append(addresses, resolver.Address{
						Addr:       addr,
						Attributes: attr.WithValue(AdvertiseAddressKey{}, true),
					})
				}
			}
		}
	}

	logger.Debug("updating state with addresses", "addresses", addresses)
	// UpdateState triggers the gRPC Balancer to re-evaluate the connection pool.
	if err := r.cc.UpdateState(resolver.State{
		Addresses: addresses,
	}); err != nil {
		logger.Error("update state fail",
			"resolve_protocol", r.protocol,
			"resolve_service", r.serviceName,
			"resolve_instance", r.instanceID,
			"resolve_registry", r.registryName,
			"err", err)
	}
}

// watch is the callback function triggered by the registry when service instances change.
func (r *clientResolver) watch(event *registry.Event) {
	logger.Debug("receive changed event", "type", event.Type, "instance", event.Instance)
	if event.Type == registry.EventTypeDelete {
		// When an instance is deleted, we refresh to get the latest valid set.
		r.cc.UpdateState(resolver.State{})
	} else {
		r.update([]*registry.Instance{event.Instance})
	}
}
