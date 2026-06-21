package registry

import (
	"testing"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/stretchr/testify/require"
)

func testInstance(id string) *Instance {
	return &Instance{DiscoverName: "local", Service: &server.Service{
		APP: runtime.APP{App: "app", Region: "region", Environment: "dev", Instance: runtime.Instance{
			ID: id, Name: "api", Version: "v1", MetaData: map[string]string{"tier": "backend"},
		}},
		Endpoints: map[string]*server.Endpoint{"grpc": {}},
	}}
}

func TestOptionsFilterInstance(t *testing.T) {
	instance := testInstance("one")
	options := NewOptions([]Option{
		WithApp("app"), WithRegion("region"), WithEnvironment("dev"), WithServiceName("api"),
		WithInstanceID("one"), WithRegistryName("local"), WithProtocol("grpc"), WithVersion("v1"),
		WithMetadata(map[string]string{"tier": "backend"}), WithPickFunc([]PickFunc{okPickFunc()}),
	})
	require.True(t, instance.canPick(options))
	require.False(t, instance.canPick(NewOptions([]Option{WithVersion("v2")})))
	require.False(t, instance.canPick(NewOptions([]Option{WithMetadata(map[string]string{"tier": "frontend"})})))
	require.True(t, instance.canPick(NewOptions([]Option{WithApp("")})))
}

func TestCacheLifecycleAndListeners(t *testing.T) {
	c := newCache(&Config{}, nil)
	var events []*Event
	options := NewOptions([]Option{WithServiceName("api"), WithWatch("watch", func(event *Event) {
		events = append(events, event)
	})})
	c.addListener(options)
	instance := testInstance("one")
	c.update([]*Instance{instance})
	require.True(t, c.isAvailable(options))
	require.Equal(t, []*Instance{instance}, c.pick(options))
	require.Len(t, events, 1)
	require.Equal(t, EventTypeUpdate, events[0].Type)

	c.delete(instance)
	require.False(t, c.isAvailable(options))
	require.Len(t, events, 2)
	require.Equal(t, EventTypeDelete, events[1].Type)
	c.removeListener("watch")
	require.NotContains(t, c.listeners, "watch")
}

func TestFailureThreshold(t *testing.T) {
	c := newCache(&Config{}, nil)
	require.Equal(t, 1, c.getFailureThreshold("new"))
	c.setFailureThreshold("one", 3)
	require.Equal(t, 3, c.getFailureThreshold("one"))
}

func TestDiscoveryOptions(t *testing.T) {
	called := false
	callback := func(*Event) { called = true }
	opts := NewDiscoveryOptions(WithDiscoveryCallback(callback))
	require.NotNil(t, opts.Callback)
	opts.Callback(nil)
	require.True(t, called)
}
