package client

import (
	"fmt"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

// ClientBuilder .
type ClientBuilder struct{}

var _ resolver.Builder = &ClientBuilder{}

// func init() {
// 	resolver.Register(&ClientBuilder{})
// }

// Build .
// target: asjard://grpc/serviceName
func (c *ClientBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	cr := &clientResolver{
		cc:          cc,
		protocol:    target.URL.Host,
		serviceName: target.Endpoint(),
	}

	cr.ResolveNow(resolver.ResolveNowOptions{})
	return cr, nil
}

func (*ClientBuilder) watch() {}

// Scheme 解析器名称
func (*ClientBuilder) Scheme() string {
	return constant.Framework
}

type clientResolver struct {
	cc          resolver.ClientConn
	protocol    string
	serviceName string
}

// Close .
func (r *clientResolver) Close() {
	registry.RemoveListener(r.listenerName())
}

// ResolveNow .
func (r *clientResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	instances := registry.PickServices(registry.WithServiceName(r.serviceName),
		registry.WithProtocol(r.protocol),
		registry.WithWatch(r.listenerName(), r.watch))
	r.update(instances)
}

func (r *clientResolver) listenerName() string {
	return fmt.Sprintf("%s_clientResolver_%s_%s",
		constant.Framework,
		r.protocol,
		r.serviceName)
}

func (r *clientResolver) update(instances []*registry.Instance) {
	// var endpoints []resolver.Endpoint
	var addresses []resolver.Address
	for _, instance := range instances {
		// var addresses []resolver.Address
		attr := attributes.New(constant.DiscoverNameKey, instance.DiscoverName)
		for mkey, mvalue := range instance.Instance.MetaData {
			attr.WithValue(mkey, mvalue)
		}
		attr = attr.WithValue(constant.ServiceAppKey, instance.Instance.App)
		attr = attr.WithValue(constant.ServiceEnvKey, instance.Instance.Environment)
		attr = attr.WithValue(constant.ServiceRegionKey, instance.Instance.Region)
		attr = attr.WithValue(constant.ServiceAZKey, instance.Instance.AZ)
		attr = attr.WithValue(constant.ServiceIDKey, instance.Instance.ID)
		attr = attr.WithValue(constant.ServiceNameKey, instance.Instance.Name)
		attr = attr.WithValue(constant.ServiceVersionKey, instance.Instance.Version)
		attr = attr.WithValue(constant.ServerProtocolKey, r.protocol)
		attr = attr.WithValue(constant.ServiceNameKey, r.serviceName)
		for name, epts := range instance.Instance.Endpoints[r.protocol] {
			for _, addr := range epts {
				addresses = append(addresses, resolver.Address{
					Addr:       addr,
					ServerName: name,
					Attributes: attr,
				})
			}
		}
		// endpoints = append(endpoints, resolver.Endpoint{
		// 	Addresses:  addresses,
		// 	Attributes: attr,
		// })
	}
	r.cc.UpdateState(resolver.State{
		Addresses: addresses,
		// Endpoints: endpoints,
	})
}

func (r *clientResolver) watch(event *registry.Event) {
	logger.Debugf("recieve changed event %v", event.Instance)
	if event.Type == registry.EventTypeDelete {
		r.cc.UpdateState(resolver.State{})
	} else {
		r.update([]*registry.Instance{event.Instance})
	}
}