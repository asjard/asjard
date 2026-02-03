/*
Package etcd implements service registration and discovery using ETCD v3.
It allows services to register their metadata with a TTL (lease) and
enables clients to watch for membership changes in the cluster.
*/
package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/stores/xetcd"
	"github.com/asjard/asjard/utils"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// NAME is the unique identifier for the ETCD registry provider.
	NAME = "etcd"
)

// Etcd represents the ETCD registration and discovery engine.
type Etcd struct {
	client           *clientv3.Client
	conf             *Config
	discoveryOptions *registry.DiscoveryOptions
}

// Config holds the parameters for the ETCD client and operation timeouts.
type Config struct {
	Client  string             `json:"client"`  // Name of the ETCD client in stores.
	Timeout utils.JSONDuration `json:"timeout"` // Global timeout for ETCD operations.
}

var (
	// Interface verification.
	_ registry.Register  = &Etcd{}
	_ registry.Discovery = &Etcd{}

	// defaultConfig provides fallback values if no configuration is found.
	defaultConfig = Config{
		Client:  xetcd.DefaultClientName,
		Timeout: utils.JSONDuration{Duration: 5 * time.Second},
	}
	newEtcd *Etcd
	newOnce sync.Once
)

func init() {
	// Add ETCD to the global list of available registry and discovery providers.
	registry.AddRegister(NAME, NewRegister)
	registry.AddDiscover(NAME, NewDiscovery)
}

// NewRegister initializes the ETCD provider for service registration.
func NewRegister() (registry.Register, error) {
	return New(nil)
}

// NewDiscovery initializes the ETCD provider for service discovery and starts the watch loop.
func NewDiscovery(options *registry.DiscoveryOptions) (registry.Discovery, error) {
	discover, err := New(options)
	if err != nil {
		return discover, err
	}
	// Start a background goroutine to listen for service changes.
	go discover.watch()
	return discover, nil
}

// New is a singleton constructor for the Etcd struct.
func New(options *registry.DiscoveryOptions) (*Etcd, error) {
	var err error
	newOnce.Do(func() {
		etcdRegistry := &Etcd{}
		err = etcdRegistry.loadConfig()
		if err != nil {
			return
		}
		// Initialize the low-level ETCD client.
		etcdRegistry.client, err = xetcd.Client(xetcd.WithClientName(etcdRegistry.conf.Client))
		if err != nil {
			return
		}
		newEtcd = etcdRegistry
	})
	if options != nil {
		newEtcd.discoveryOptions = options
	}
	return newEtcd, err
}

// GetAll retrieves a full snapshot of all active service instances currently in ETCD.
func (e *Etcd) GetAll() ([]*registry.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.conf.Timeout.Duration)
	defer cancel()

	// Fetch all keys under the application prefix.
	resp, err := e.client.Get(ctx, e.prefix(), clientv3.WithPrefix())
	if err != nil {
		return []*registry.Instance{}, err
	}

	instances := make([]*registry.Instance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var service server.Service
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			return instances, err
		}
		instances = append(instances, &registry.Instance{
			DiscoverName: NAME,
			Service:      &service,
		})
	}
	return instances, nil
}

// Name returns the provider identifier "etcd".
func (e *Etcd) Name() string {
	return NAME
}

// Registe uploads the service instance details to ETCD with a 5-second TTL lease.
// It also manages the KeepAlive heartbeats to maintain the registration.
func (e *Etcd) Registe(instance *server.Service) error {
	logger.Debug("register instance into etcd", "instance", instance)
	b, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("register instance fail[%s]", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a lease with 5 seconds of life.
	lease := clientv3.NewLease(e.client)
	grant, err := lease.Grant(ctx, 5)
	if err != nil {
		return err
	}

	// Put the instance JSON into ETCD attached to the lease.
	if _, err := e.client.Put(ctx,
		e.registerKey(instance),
		string(b),
		clientv3.WithLease(grant.ID)); err != nil {
		return fmt.Errorf("register instance fail[%s]", err)
	}

	// Start automatic heartbeat.
	leaseChan, err := lease.KeepAlive(context.Background(), grant.ID)
	if err != nil {
		return err
	}

	// Monitoring goroutine for the lease status.
	go func() {
		for {
			select {
			case resp := <-leaseChan:
				// If leaseChan is closed or nil, the lease is lost.
				if resp == nil {
					logger.Error("keepalive fail")
					// Attempt to re-register indefinitely every 3 seconds.
					for {
						logger.Debug("start keepalive instance")
						if err := e.Registe(instance); err != nil {
							logger.Error("keepalive instance fail", "err", err)
						} else {
							logger.Debug("rekeepalive instance success")
							return
						}
						time.Sleep(3 * time.Second)
					}
				}
			}
		}
	}()
	return nil
}

// Remove manually deletes the instance key from ETCD (e.g., during graceful shutdown).
func (e *Etcd) Remove(instance *server.Service) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := e.client.Delete(ctx, e.registerKey(instance)); err != nil {
		logger.Error("delete instance fail", "err", err)
	}
}

// watch listens for ETCD events (PUT/DELETE) and triggers callbacks for the load balancer.
func (e *Etcd) watch() {
	logger.Debug("watch instance from etcd")
	watchChan := e.client.Watch(context.Background(), e.prefix(), clientv3.WithPrefix())
	for resp := range watchChan {
		for _, event := range resp.Events {
			logger.Debug("instance from etcd updated",
				"key", string(event.Kv.Key),
				"event", event.Type.String())

			callbackEvent := &registry.Event{
				Instance: &registry.Instance{
					DiscoverName: e.Name(),
				},
			}

			switch event.Type {
			case mvccpb.PUT:
				// An instance was added or updated.
				callbackEvent.Type = registry.EventTypeUpdate
				var service server.Service
				if err := json.Unmarshal(event.Kv.Value, &service); err != nil {
					logger.Error("unmarshal service fail",
						"key", string(event.Kv.Key),
						"value", string(event.Kv.Value),
						"err", err)
					continue
				}
				callbackEvent.Instance.Service = &service

			case mvccpb.DELETE:
				// An instance was removed or its lease expired.
				callbackEvent.Type = registry.EventTypeDelete
				keyList := strings.Split(string(event.Kv.Key), "/")
				if len(keyList) > 0 {
					callbackEvent.Instance.Service = &server.Service{
						APP: runtime.APP{
							Instance: runtime.Instance{
								ID: keyList[len(keyList)-1], // Extract ID from the key.
							},
						},
					}
				}
			}
			// Notify the framework's resolver/load balancer of the change.
			e.discoveryOptions.Callback(callbackEvent)
		}
	}
	logger.Debug("watch exit")
}

func (e *Etcd) loadConfig() error {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.registry.etcd", &conf); err != nil {
		return err
	}
	e.conf = &conf
	return nil
}

// registerKey generates the full ETCD path for a service: /app/instances/env/serviceName/instanceID.
func (e *Etcd) registerKey(instance *server.Service) string {
	return fmt.Sprintf("%s/%s/%s", e.prefix(), instance.Instance.Name, instance.Instance.ID)
}

// prefix builds the root path for service discovery based on app name and environment.
func (e *Etcd) prefix() string {
	return "/" + runtime.GetAPP().App + "/instances/" + runtime.GetAPP().Environment
}
