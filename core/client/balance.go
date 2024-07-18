package client

import (
	"github.com/asjard/asjard/core/logger"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const (
	// DefaultBalanceRoundRobin 默认负载均衡器roundrobin
	DefaultBalanceRoundRobin = "roundRobin"
	// HeaderRequestSource 请求源服务
	HeaderRequestSource = "x-request-source"
	// HeaderRequestDest 请求目的地
	HeaderRequestDest = "x-request-dest"
	// HeaderRequestApp 请求应用
	HeaderRequestApp = "x-request-app"
)

type SubConn struct {
	Address resolver.Address
	Conn    balancer.SubConn
}

// NewBalancerPicker .
type NewBalancerPicker func(scs map[balancer.SubConn]base.SubConnInfo) Picker

// 负载均衡器列表
var balancers = make(map[string]NewBalancerPicker)

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
		logger.Warn("loadbalance name is empty, set to default",
			"default", DefaultBalanceRoundRobin)
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
		logger.Warn("loadbalance not found, set to default",
			"loadbalance", balanceName,
			"default", DefaultBalanceRoundRobin)
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
	return NewPicker(b.newPicker, scs)
}
