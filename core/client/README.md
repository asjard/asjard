> 客户端管理实现, 当前只实现了grpc客户端， 其他协议请按照grpc客户端实现

### 客户端功能列表

- 负载均衡
- 拦截器


### 自定义负载均衡实现

```go
func init() {
	AddBalancer("roundRobin", NewRoundRobinPicker)
}

/ NewRoundRobinPicker .
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

type subConn struct {
	address resolver.Address
	conn    balancer.SubConn
}

// RoundRobinPicker 轮询负载均衡
type RoundRobinPicker struct {
	scs  []*subConn
	next uint32
}

// Pick 负载选择
func (r *RoundRobinPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	n := uint32(len(r.scs))
	if n == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	next := atomic.AddUint32(&r.next, 1) - 1
	sc := r.scs[next%n]
	return balancer.PickResult{
		SubConn: sc.conn,
		Done:    func(info balancer.DoneInfo) {},
	}, nil
}
```
