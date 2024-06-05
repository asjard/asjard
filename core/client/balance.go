package client

import (
	"sync/atomic"

	"github.com/asjard/asjard/core/logger"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const (
	// DefaultBalanceRoundRobin 默认负载均衡器roundrobin
	DefaultBalanceRoundRobin = "roundRobin"
)

// NewBalancerPicker .
type NewBalancerPicker func(scs map[balancer.SubConn]base.SubConnInfo) balancer.Picker

// 负载均衡器列表
var balancers = make(map[string]NewBalancerPicker)

func init() {
	AddBalancer(DefaultBalanceRoundRobin, NewRoundRobinPicker)
}

// AddBalancer 添加负载均衡器
func AddBalancer(name string, newPicker NewBalancerPicker) {
	balancers[name] = newPicker
}

// BalanceBuilder 负载均衡构造器
type BalanceBuilder struct {
	newPicker NewBalancerPicker
}

var _ base.PickerBuilder = &BalanceBuilder{}

// NewBalanceBuilder .
func NewBalanceBuilder(balanceName string) balancer.Builder {
	if balanceName == "" {
		logger.Warnf("loadbalance name is empty, set to default %s", DefaultBalanceRoundRobin)
		balanceName = DefaultBalanceRoundRobin
	}
	balanceBuilder := &BalanceBuilder{}
	exist := false
	for pickerName, newPicker := range balancers {
		if pickerName == balanceName {
			balanceBuilder.newPicker = newPicker
			exist = true
			break
		}
	}
	if !exist {
		logger.Warnf("loadbalance %s not found, set to default %s", balanceName, DefaultBalanceRoundRobin)
		balanceBuilder.newPicker = NewRoundRobinPicker
	}
	return base.NewBalancerBuilder(balanceName, balanceBuilder, base.Config{HealthCheck: true})
}

// Build .
func (b *BalanceBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	scs := make(map[balancer.SubConn]base.SubConnInfo)
	for sc, sinfo := range info.ReadySCs {
		scs[sc] = sinfo
	}
	return b.newPicker(scs)
}

// NewRoundRobinPicker .
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

// Pick .
func (r *RoundRobinPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	n := uint32(len(r.scs))
	if n == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// TODO 整形溢出
	next := atomic.AddUint32(&r.next, 1) - 1
	sc := r.scs[next%n]
	return balancer.PickResult{
		SubConn: sc.conn,
		Done:    func(info balancer.DoneInfo) {},
	}, nil
}
