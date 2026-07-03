package consul

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/utils"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestConsulNameAndConfig(t *testing.T) {
	c := &Consul{}
	require.Equal(t, NAME, c.Name())
	require.NoError(t, c.loadConfig())
	require.NotNil(t, c.conf)
}

func TestInstanceWatchEvents(t *testing.T) {
	app := runtime.APP{App: "app", Instance: runtime.Instance{ID: "one", Name: "api"}}
	endpoints := map[string]*server.Endpoint{"grpc": {Advertise: []string{"service:9000"}}}
	appJSON, err := json.Marshal(app)
	require.NoError(t, err)
	endpointJSON, err := json.Marshal(endpoints)
	require.NoError(t, err)

	var events []*registry.Event
	c := &Consul{discoveryOptions: &registry.DiscoveryOptions{Callback: func(event *registry.Event) { events = append(events, event) }}}
	w := &instanceWatch{instances: make(map[string]uint64), c: c, service: "api"}
	w.handler(0, []*api.ServiceEntry{{Service: &api.AgentService{
		ID: "one", ModifyIndex: 1, Meta: map[string]string{"app_detail": string(appJSON), "endpoints": string(endpointJSON)},
	}}})
	require.Len(t, events, 1)
	require.Equal(t, registry.EventTypeCreate, events[0].Type)
	require.Equal(t, "one", events[0].Instance.Service.Instance.ID)

	w.handler(0, []*api.ServiceEntry{})
	require.Len(t, events, 2)
	require.Equal(t, registry.EventTypeDelete, events[1].Type)
	require.Equal(t, "one", events[1].Instance.Service.Instance.ID)
}

func TestConsulRegistryIntegration(t *testing.T) {
	apiConfig := api.DefaultConfig()
	apiConfig.Address = "127.0.0.1:8500"
	client, err := api.NewClient(apiConfig)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		_, err := client.Status().Leader()
		return err == nil
	}, 30*time.Second, 500*time.Millisecond)

	provider := &Consul{client: client, conf: &Config{Timeout: utils.JSONDuration{Duration: 5 * time.Second}}, exit: make(chan struct{})}
	service := &server.Service{APP: runtime.GetAPP(), Endpoints: map[string]*server.Endpoint{"grpc": {Advertise: []string{"127.0.0.1:9000"}}}}
	service.Instance.ID = fmt.Sprintf("asjard-test-%d", time.Now().UnixNano())
	require.NoError(t, provider.Registe(service))
	t.Cleanup(func() { provider.Remove(service) })
	require.Eventually(t, func() bool {
		instances, err := provider.GetAll()
		if err != nil {
			return false
		}
		for _, instance := range instances {
			if instance.Service.Instance.ID == service.Instance.ID {
				return true
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond)
}
