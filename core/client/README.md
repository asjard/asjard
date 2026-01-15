### Custom Load Balancer Implementation

```go
func init() {
    // Registers the custom balancer with the global registry.
    AddBalancer("roundRobin", NewRoundRobinPicker)
}

// NewRoundRobinPicker initializes a new Round Robin picker instance.
func NewRoundRobinPicker(scs map[balancer.SubConn]base.SubConnInfo) balancer.Picker {
    var subConns []*subConn
    for conn, info := range scs {
        subConns = append(subConns, &subConn{
            address: info.Address,
            conn:    conn,
        })
    }
    return &RoundRobinPicker{
        scs: subConns,
    }
}

// subConn wraps the gRPC SubConn with its resolved address metadata.
type subConn struct {
    address resolver.Address
    conn    balancer.SubConn
}

// RoundRobinPicker implements the Round Robin load balancing strategy.
type RoundRobinPicker struct {
    scs  []*subConn
    next uint32
}

// Pick selects a sub-connection from the available pool using a Round Robin algorithm.
func (r *RoundRobinPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
    n := uint32(len(r.scs))
    if n == 0 {
        return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
    }

    // Atomically increment the counter to ensure thread-safe selection.
    next := atomic.AddUint32(&r.next, 1) - 1
    sc := r.scs[next%n]

    return balancer.PickResult{
        SubConn: sc.conn,
        // Done is called when the RPC is completed.
        Done:    func(info balancer.DoneInfo) {},
    }, nil
}

// Name returns the unique identifier for this load balancing strategy.
func (r *RoundRobinPicker) Name() string {
    return "roundRobin"
}
```
