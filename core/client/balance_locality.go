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
	// HeaderRequestRegion identifies the target geographic region for the request.
	HeaderRequestRegion = "x-request-region"
	// HeaderRequestAz identifies the target availability zone for the request.
	HeaderRequestAz = "x-request-az"
	// LocalityRoundRobinName is the unique identifier for this balancer strategy.
	LocalityRoundRobinName = "localityRoundRobin"
)

// LocalityRoundRobinPicker implements a locality-aware load balancing strategy.
// It prioritizes sub-connections in the same Availability Zone (AZ) or Region
// to reduce cross-data-center latency.
type LocalityRoundRobinPicker struct {
	*PickerBase
	scs  []*SubConn
	next uint32
	app  runtime.APP
}

func init() {
	// Automatically register the locality-aware picker to the client balancer registry.
	AddBalancer(LocalityRoundRobinName, NewLocalityRoundRobinPicker)
}

// NewLocalityRoundRobinPicker creates a new instance of the locality-aware picker.
// It maps gRPC sub-connections to the internal SubConn wrapper containing metadata.
func NewLocalityRoundRobinPicker(scs map[balancer.SubConn]base.SubConnInfo) Picker {
	subConns := make([]*SubConn, 0, len(scs))
	for conn, info := range scs {
		subConns = append(subConns, &SubConn{
			Address: info.Address,
			Conn:    conn,
		})
	}
	app := runtime.GetAPP()
	return &LocalityRoundRobinPicker{
		PickerBase: NewPickerBase(app),
		scs:        subConns,
		app:        app,
	}
}

// Pick selects a sub-connection based on locality priority:
// 1. Match requested AZ (Availability Zone).
// 2. If no match, match requested Region.
// 3. Fallback to shareable instances if specific locality is unreachable.
func (l *LocalityRoundRobinPicker) Pick(info balancer.PickInfo) (*PickResult, error) {
	var requestRegion, requestAz string
	// Extract locality hints from gRPC metadata (incoming context).
	md, ok := metadata.FromIncomingContext(info.Ctx)
	if ok {
		if regions := md.Get(HeaderRequestRegion); len(regions) > 0 {
			requestRegion = regions[0]
		}
		if azs := md.Get(HeaderRequestAz); len(azs) > 0 {
			requestAz = azs[0]
		}
	}

	// Multi-stage filtering: Filter by Region first, then refine by AZ.
	picks := l.pick(requestAz, l.app.AZ, l.isSameAz,
		l.pick(requestRegion, l.app.Region, l.isSameRegion, l.scs))

	n := uint32(len(picks))
	if n == 0 {
		return nil, balancer.ErrNoSubConnAvailable
	}

	// Perform standard Round Robin across the filtered subset of connections.
	next := atomic.AddUint32(&l.next, 1) - 1
	sc := picks[next%n]

	return &PickResult{
		SubConn:       sc,
		RequestRegion: requestRegion,
		RequestAz:     requestAz,
	}, nil
}

// Name returns the identifier of the balancer.
func (l *LocalityRoundRobinPicker) Name() string {
	return LocalityRoundRobinName
}

// pick filters a list of sub-connections based on the provided locality criteria.
// It includes fallback logic to find shareable instances if the specific locality match fails.
func (l LocalityRoundRobinPicker) pick(request, current string, isEqual func(request string, sc *SubConn) bool, scs []*SubConn) []*SubConn {
	if len(scs) == 0 {
		return []*SubConn{}
	}

	// If no specific request locality is provided, default to the current application's locality.
	if request == "" {
		return l.pick(current, current, isEqual, scs)
	}

	picks := make([]*SubConn, 0, len(scs))
	for _, sc := range scs {
		if isEqual(request, sc) && l.CanReachable(sc) {
			picks = append(picks, sc)
		}
	}

	// Fallback mechanism: if no instances match the requested locality,
	// try to find shareable instances in the current application's locality.
	if len(picks) == 0 && request != current {
		logger.Debug("no conns in request locality, falling back to current",
			"request", request, "current", current)
		return l.pick(current, current, l.isShareable, scs)
	}
	return picks
}

// isSameRegion checks if the sub-connection's instance belongs to the specified region.
func (l LocalityRoundRobinPicker) isSameRegion(region string, sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if ok && instance.Service.Region == region {
		return true
	}
	return false
}

// isSameAz checks if the sub-connection's instance belongs to the specified availability zone.
func (l LocalityRoundRobinPicker) isSameAz(az string, sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if ok && instance.Service.AZ == az {
		return true
	}
	return false
}

// isShareable determines if a sub-connection is eligible for traffic sharing
// when local matches are unavailable.
func (l LocalityRoundRobinPicker) isShareable(_ string, sc *SubConn) bool {
	return l.Shareable(sc)
}
