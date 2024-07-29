package client

import (
	"sync/atomic"

	"github.com/asjard/asjard/core/logger"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

// RoundRobinPicker 轮询负载均衡
type RoundRobinPicker struct {
	*PickerBase
	// 所有连接列表
	scs  []*SubConn
	next uint32
}

func init() {
	AddBalancer(DefaultBalanceRoundRobin, NewRoundRobinPicker)
}

// NewRoundRobinPicker .
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
		PickerBase: NewPickerBase(),
		scs:        subConns,
	}
}

// Pick 负载选择
func (r *RoundRobinPicker) Pick(info balancer.PickInfo) (*PickResult, error) {
	n := uint32(len(r.scs))
	if n == 0 {
		return nil, balancer.ErrNoSubConnAvailable
	}
	picks := make([]*SubConn, 0, len(r.scs))
	for _, sc := range r.scs {
		if r.CanReachable(sc) {
			picks = append(picks, sc)
		}
	}
	n = uint32(len(picks))
	next := atomic.AddUint32(&r.next, 1) - 1
	sc := picks[next%n]
	return &PickResult{
		SubConn: sc,
	}, nil
}

func (r *RoundRobinPicker) Name() string {
	return DefaultBalanceRoundRobin
}
