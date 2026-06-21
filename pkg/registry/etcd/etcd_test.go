package etcd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/utils"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestEtcdMetadataAndKeys(t *testing.T) {
	e := &Etcd{}
	require.Equal(t, NAME, e.Name())
	require.NoError(t, e.loadConfig())
	require.NotNil(t, e.conf)
	prefix := e.prefix()
	require.True(t, strings.HasPrefix(prefix, "/"))
	service := &server.Service{}
	service.Instance.Name = "api"
	service.Instance.ID = "instance"
	require.Equal(t, prefix+"/api/instance", e.registerKey(service))
}

func TestEtcdRegistryIntegration(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, DialTimeout: 5 * time.Second})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.Eventually(t, func() bool {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err := client.Status(ctx, "127.0.0.1:2379")
		return err == nil
	}, 30*time.Second, 500*time.Millisecond)

	provider := &Etcd{client: client, conf: &Config{Timeout: utils.JSONDuration{Duration: 5 * time.Second}}}
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
