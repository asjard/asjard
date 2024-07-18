package client

import (
	"sync/atomic"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
)

const (
	HeaderRequestRegion = "x-request-region"
	HeaderRequestAz     = "x-request-az"
)

// LocalityRoundRobinPicker 就近权重负载均衡
type LocalityRoundRobinPicker struct {
	*PickerBase
	scs  []*SubConn
	next uint32
	app  runtime.APP
}

func init() {
	AddBalancer("localityRoundRobin", NewLocalityRoundRobinPicker)
}

// NewLocalityWeightPicker 就近权重负载均衡初始化
func NewLocalityRoundRobinPicker(scs map[balancer.SubConn]base.SubConnInfo) Picker {
	logger.Debug("locality round robin picker")
	subConns := make([]*SubConn, 0, len(scs))
	for conn, info := range scs {
		subConns = append(subConns, &SubConn{
			Address: info.Address,
			Conn:    conn,
		})
	}
	return &LocalityRoundRobinPicker{
		PickerBase: NewPickerBase(),
		scs:        subConns,
		app:        runtime.GetAPP(),
	}
}

// Pick 负载均衡选择
// 同region,az优先
func (l *LocalityRoundRobinPicker) Pick(info balancer.PickInfo) (*SubConn, error) {
	var requestRegion, requestAz string
	md, ok := metadata.FromIncomingContext(info.Ctx)
	if ok {
		if regions := md.Get(HeaderRequestRegion); len(regions) > 0 {
			requestRegion = regions[0]
		}
		if azs := md.Get(HeaderRequestAz); len(azs) > 0 {
			requestAz = azs[0]
		}
	}
	picks := l.pick(requestAz, l.app.AZ, l.isSameAz,
		l.pick(requestRegion, l.app.Region, l.isSameRegion, l.scs))

	n := uint32(len(picks))
	if n == 0 {
		return nil, balancer.ErrNoSubConnAvailable
	}
	next := atomic.AddUint32(&l.next, 1) - 1
	sc := picks[next%n]
	return sc, nil
}

// 优先选择request下的
// 如果request下为空，则选择current下的
// 如果current下为空，则获被共享的
func (l LocalityRoundRobinPicker) pick(request, current string, isEqual func(request string, sc *SubConn) bool, scs []*SubConn) []*SubConn {
	if len(scs) == 0 {
		return []*SubConn{}
	}
	if request == "" {
		return l.pick(current, current, isEqual, scs)
	}
	picks := make([]*SubConn, 0, len(scs))
	for _, sc := range scs {
		if request != current && isEqual(request, sc) && l.CanReachable(sc) {
			picks = append(picks, sc)
		}
	}
	if len(picks) == 0 {
		if request != current {
			logger.Debug("request is empty")
			// 获取current下共享的
			return l.pick(current, current, isEqual, scs)
		}
		for _, sc := range scs {
			if l.Shareable(sc) && l.CanReachable(sc) {
				picks = append(picks, sc)
			}
		}
	}
	return picks
}

func (l LocalityRoundRobinPicker) isSameRegion(region string, sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	logger.Debug("----same-region--", "request", region, "current", instance.Service.Region, "sc", sc)
	if ok && instance.Service.Region == region {
		return true
	}
	return false
}

func (l LocalityRoundRobinPicker) isSameAz(az string, sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	logger.Debug("----same-az--", "request", az, "current", instance.Service.AZ, "sc", sc)
	if ok && instance.Service.AZ == az {
		return true
	}
	return false
}
