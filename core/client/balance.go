package client

import (
	"sync/atomic"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils/cast"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
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

// Pick 负载选择
// TODO
// 优先同app，region,az
// 优先选择同区域
// 如果垮区域应优先使用advertise地址
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
		Metadata: metadata.New(map[string]string{
			HeaderRequestSource: runtime.GetInstance().Name,
			HeaderRequestDest:   cast.ToString(sc.address.Attributes.Value(constant.ServiceNameKey)),
			HeaderRequestApp:    cast.ToString(sc.address.Attributes.Value(constant.ServiceAppKey)),
		}),
	}, nil
}
