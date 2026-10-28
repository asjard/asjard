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
	conf             *Config
	discoveryOptions *registry.DiscoveryOptions

	mu     sync.Mutex
	leases map[string]*instanceLease
}

type instanceLease struct {
	cancel context.CancelFunc
	id     clientv3.LeaseID
	gen    uint64
}

// Config holds the parameters for the ETCD client and operation timeouts.
type Config struct {
	Client  string             `json:"client"`  // Name of the ETCD client in stores.
	Timeout utils.JSONDuration `json:"timeout"` // Global timeout for ETCD operations.
	TTL     utils.JSONDuration `json:"ttl"`     // Lease TTL used by KeepAlive.
}

var (
	// Interface verification.
	_ registry.Register  = &Etcd{}
	_ registry.Discovery = &Etcd{}

	// defaultConfig provides fallback values if no configuration is found.
	defaultConfig = Config{
		Client:  xetcd.DefaultClientName,
		Timeout: utils.JSONDuration{Duration: 5 * time.Second},
		TTL:     utils.JSONDuration{Duration: 30 * time.Second},
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
		_, err = xetcd.Client(xetcd.WithClientName(etcdRegistry.conf.Client))
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

func (e *Etcd) etcdClient() (*clientv3.Client, error) {
	client, err := xetcd.Client(xetcd.WithClientName(e.conf.Client))
	if err != nil {
		return nil, nil
	}
	return client, nil
}

func (e *Etcd) opTimeout() time.Duration {
	if e.conf == nil || e.conf.Timeout.Duration <= 0 {
		return 5 * time.Second
	}
	return e.conf.Timeout.Duration
}

func (e *Etcd) ttlSeconds() int64 {
	if e.conf == nil || e.conf.TTL.Duration <= 0 {
		return int64(defaultConfig.TTL.Duration.Seconds())
	}
	sec := int64(e.conf.TTL.Duration.Seconds())
	if sec < 1 {
		return 1
	}
	return sec
}

// GetAll retrieves a full snapshot of all active service instances currently in ETCD.
func (e *Etcd) GetAll() ([]*registry.Instance, error) {
	client, err := e.etcdClient()
	if err != nil {
		return []*registry.Instance{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), e.opTimeout())
	defer cancel()

	// Fetch all keys under the application prefix.
	resp, err := client.Get(ctx, e.prefix(), clientv3.WithPrefix())
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

// Registe uploads the service instance details to ETCD with a TTL lease.
// It also manages the KeepAlive heartbeats to maintain the registration.
func (e *Etcd) Registe(instance *server.Service) error {
	logger.Debug("register instance into etcd", "instance", instance)
	client, err := e.etcdClient()
	if err != nil {
		return err
	}
	b, err := json.Marshal(instance)
	if err != nil {
		return fmt.Errorf("register instance fail[%s]", err)
	}

	key := e.registerKey(instance)
	gen := e.nextLeaseGen(key)

	ctx, cancel := context.WithTimeout(context.Background(), e.opTimeout())
	defer cancel()

	grant, err := client.Grant(ctx, e.ttlSeconds())
	if err != nil {
		return err
	}

	if _, err := client.Put(ctx, key, string(b), clientv3.WithLease(grant.ID)); err != nil {
		return fmt.Errorf("register instance fail[%s]", err)
	}

	kaCtx, kaCancel := context.WithCancel(context.Background())
	leaseChan, err := client.KeepAlive(kaCtx, grant.ID)
	if err != nil {
		kaCancel()
		return err
	}

	e.storeLease(key, &instanceLease{cancel: kaCancel, id: grant.ID, gen: gen})
	go e.watchLease(instance, key, gen, leaseChan)
	return nil
}

func (e *Etcd) watchLease(instance *server.Service, key string, gen uint64, leaseChan <-chan *clientv3.LeaseKeepAliveResponse) {
	watchLease(leaseChan, func() {
		if !e.shouldReregister(key, gen) {
			return
		}
		logger.Error("keepalive channel closed")
		e.reregiste(instance)
	})
}

func watchLease(leaseChan <-chan *clientv3.LeaseKeepAliveResponse, onLost func()) {
	for resp := range leaseChan {
		if resp == nil {
			onLost()
			return
		}
	}
	onLost()
}

func (e *Etcd) reregiste(instance *server.Service) {
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

// Remove manually deletes the instance key from ETCD (e.g., during graceful shutdown).
func (e *Etcd) Remove(instance *server.Service) {
	key := e.registerKey(instance)
	e.stopLease(key)
	client, err := e.etcdClient()
	if err != nil {
		logger.Error("delete instance fail", "err", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), e.opTimeout())
	defer cancel()
	if _, err := client.Delete(ctx, key); err != nil {
		logger.Error("delete instance fail", "err", err)
	}
}

func (e *Etcd) nextLeaseGen(key string) uint64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leases == nil {
		e.leases = make(map[string]*instanceLease)
	}
	gen := uint64(1)
	if old, ok := e.leases[key]; ok {
		gen = old.gen + 1
		if old.cancel != nil {
			old.cancel()
		}
	}
	e.leases[key] = &instanceLease{gen: gen}
	return gen
}

func (e *Etcd) storeLease(key string, lease *instanceLease) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.leases == nil {
		e.leases = make(map[string]*instanceLease)
	}
	e.leases[key] = lease
}

func (e *Etcd) shouldReregister(key string, gen uint64) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	lease, ok := e.leases[key]
	return ok && lease.gen == gen
}

func (e *Etcd) stopLease(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if lease, ok := e.leases[key]; ok {
		if lease.cancel != nil {
			lease.cancel()
		}
		delete(e.leases, key)
	}
}

// watch listens for ETCD events (PUT/DELETE) and triggers callbacks for the load balancer.
func (e *Etcd) watch() {
	logger.Debug("watch instance from etcd")
	for {
		client, err := e.etcdClient()
		if err != nil {
			logger.Error("watch instance from etcd fail", "err", err)
			time.Sleep(time.Second)
			continue
		}
		ctx, cancel := context.WithCancel(context.Background())
		e.consumeWatch(client.Watch(ctx, e.prefix(), clientv3.WithPrefix()))
		cancel()
		logger.Warn("watch instance from etcd closed, restarting")
		time.Sleep(time.Second)
	}
}

func (e *Etcd) consumeWatch(watchChan clientv3.WatchChan) {
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
