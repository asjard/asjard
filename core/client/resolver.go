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

// AddressAttrKey 空的结构体用来作为address attr的key
type AddressAttrKey struct{}

// ListenAddressKey 空的结构用来作为标记监听地址
type ListenAddressKey struct{}

// AdvertiseAddressKey 空的结构体用来标记广播地址
type AdvertiseAddressKey struct{}

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
	app := runtime.GetAPP()
	instances := registry.PickServices(registry.WithServiceName(r.serviceName),
		registry.WithProtocol(r.protocol),
		registry.WithEnvironment(app.Environment),
		registry.WithApp(app.App),
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
		attr := attributes.New(AddressAttrKey{}, instance)
		endpoint, ok := instance.Service.GetEndpoint(r.protocol)
		if ok {
			if len(endpoint.Listen) != 0 {
				for _, addr := range endpoint.Listen {
					addresses = append(addresses, resolver.Address{
						Addr:       addr,
						Attributes: attr.WithValue(ListenAddressKey{}, true),
					})
				}
			}
			if len(endpoint.Advertise) != 0 {
				for _, addr := range endpoint.Advertise {
					addresses = append(addresses, resolver.Address{
						Addr:       addr,
						Attributes: attr.WithValue(AdvertiseAddressKey{}, true),
					})
				}
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
