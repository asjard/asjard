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

// ClientBuilder .
type ClientBuilder struct{}

var _ resolver.Builder = &ClientBuilder{}

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

// ResolveNow 从服务发现中心获取服务列表
func (r *clientResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	instances := registry.PickServices(registry.WithServiceName(r.serviceName),
		registry.WithProtocol(r.protocol),
		registry.WithEnvironment(runtime.GetAPP().Environment),
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
	var addresses []resolver.Address
	for _, instance := range instances {
		attr := attributes.New(constant.DiscoverNameKey, instance.DiscoverName)
		for mkey, mvalue := range instance.Instance.Instance.MetaData {
			attr.WithValue(mkey, mvalue)
		}
		attr = attr.WithValue(constant.ServiceAppKey, instance.Instance.App)
		attr = attr.WithValue(constant.ServiceEnvKey, instance.Instance.Environment)
		attr = attr.WithValue(constant.ServiceRegionKey, instance.Instance.Region)
		attr = attr.WithValue(constant.ServiceAZKey, instance.Instance.AZ)
		attr = attr.WithValue(constant.ServiceIDKey, instance.Instance.Instance.ID)
		attr = attr.WithValue(constant.ServiceNameKey, instance.Instance.Instance.Name)
		attr = attr.WithValue(constant.ServiceVersionKey, instance.Instance.Instance.Version)
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
	}
	r.cc.UpdateState(resolver.State{
		Addresses: addresses,
	})
}

func (r *clientResolver) watch(event *registry.Event) {
	logger.Debug("recieve changed event",
		"event", event.Instance)
	if event.Type == registry.EventTypeDelete {
		r.cc.UpdateState(resolver.State{})
	} else {
		r.update([]*registry.Instance{event.Instance})
	}
}
