## 配置

```yaml
## client configurations
clients:
  ## client loadbalance, default: localityRoundRobin
  # loadbalance: "localityRoundRobin"
  ## grpc client configuration
  grpc:
    ## grpc client loadbalance
    # loadbalance: ""
```

## 已实现负载均衡器

- [本地优先轮询](balance-locality.md)
- [轮询](balance-roundrobin.md)

## 自定义balance

- 实现如下接口

```go
// Picker defines the interface that all custom load balancers must implement.
// It is a higher-level abstraction over gRPC's native balancer.Picker.
type Picker interface {
	// Pick selects a backend connection based on the pick information.
	Pick(info balancer.PickInfo) (*PickResult, error)

	// Name returns the name of the load balancing strategy (used in logs and headers).
	Name() string
}
```

## 实现

```go
package client

import (
	"sync/atomic"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

// RoundRobinPicker implements a standard weighted-less round-robin load balancing algorithm.
// It cycles through available sub-connections sequentially to ensure even traffic distribution.
type RoundRobinPicker struct {
	*PickerBase
	// scs is the list of all available sub-connections.
	scs []*SubConn
	// next is an atomic counter used to track the index of the next connection to pick.
	next uint32
}

func init() {
	// Register the round-robin picker as the default balancer strategy.
	AddBalancer(DefaultBalanceRoundRobin, NewRoundRobinPicker)
}

// NewRoundRobinPicker creates a new instance of the RoundRobinPicker.
// It initializes the sub-connection list and sets up the base picker with the current application context.
func NewRoundRobinPicker(scs map[balancer.SubConn]base.SubConnInfo) Picker {
	logger.Debug("new roundrobin picker")
	subConns := make([]*SubConn, 0, len(scs))
	for conn, info := range scs {
		subConns = append(subConns, &SubConn{
			Address: info.Address,
			Conn:    conn,
		})
	}
	return &RoundRobinPicker{
		PickerBase: NewPickerBase(runtime.GetAPP()),
		scs:        subConns,
	}
}

// Pick selects a sub-connection using the round-robin algorithm.
// It filters connections based on their reachability status before picking.
func (r *RoundRobinPicker) Pick(info balancer.PickInfo) (*PickResult, error) {
	n := uint32(len(r.scs))
	if n == 0 {
		return nil, balancer.ErrNoSubConnAvailable
	}

	// Filter connections to find those that are currently reachable/healthy.
	picks := make([]*SubConn, 0, len(r.scs))
	for _, sc := range r.scs {
		if r.CanReachable(sc) {
			picks = append(picks, sc)
		}
	}

	n = uint32(len(picks))
	if n == 0 {
		return nil, balancer.ErrNoSubConnAvailable
	}

	// Increment the counter atomically and use modulo to select the next connection.
	next := atomic.AddUint32(&r.next, 1) - 1
	sc := picks[next%n]

	return &PickResult{
		SubConn: sc,
	}, nil
}

// Name returns the identifier for this load balancing strategy.
func (r *RoundRobinPicker) Name() string {
	return DefaultBalanceRoundRobin
}

```
