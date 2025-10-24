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
	HeaderRequestRegion    = "x-request-region"
	HeaderRequestAz        = "x-request-az"
	LocalityRoundRobinName = "localityRoundRobin"
)

// LocalityRoundRobinPicker locality round robin loadbalance
type LocalityRoundRobinPicker struct {
	*PickerBase
	scs  []*SubConn
	next uint32
	app  runtime.APP
}

func init() {
	AddBalancer(LocalityRoundRobinName, NewLocalityRoundRobinPicker)
}

// NewLocalityRoundRobinPicker create a new locality round robind balance
func NewLocalityRoundRobinPicker(scs map[balancer.SubConn]base.SubConnInfo) Picker {
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

// Pick a result with policy locality round robin.
func (l *LocalityRoundRobinPicker) Pick(info balancer.PickInfo) (*PickResult, error) {
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
	return &PickResult{
		SubConn:       sc,
		RequestRegion: requestRegion,
		RequestAz:     requestAz,
	}, nil
}

func (l *LocalityRoundRobinPicker) Name() string {
	return LocalityRoundRobinName
}

func (l LocalityRoundRobinPicker) pick(request, current string, isEqual func(request string, sc *SubConn) bool, scs []*SubConn) []*SubConn {
	if len(scs) == 0 {
		return []*SubConn{}
	}
	if request == "" {
		return l.pick(current, current, isEqual, scs)
	}
	picks := make([]*SubConn, 0, len(scs))
	for _, sc := range scs {
		if isEqual(request, sc) && l.CanReachable(sc) {
			picks = append(picks, sc)
		}
	}
	if len(picks) == 0 && request != current {
		logger.Debug("no conns in request", "request", request, "current", current)
		// pick an shareable instance from current
		return l.pick(current, current, l.isShareable, scs)
	}
	return picks
}

func (l LocalityRoundRobinPicker) isSameRegion(region string, sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if ok && instance.Service.Region == region {
		return true
	}
	return false
}

func (l LocalityRoundRobinPicker) isSameAz(az string, sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if ok && instance.Service.AZ == az {
		return true
	}
	return false
}

func (l LocalityRoundRobinPicker) isShareable(_ string, sc *SubConn) bool {
	return l.Shareable(sc)
}
