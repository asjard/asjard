/*
Package client provides gRPC client-side utilities, specifically focused on
custom load balancing and picker orchestration.
*/
package client

import (
	"github.com/asjard/asjard/core/logger"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const (
	// DefaultBalanceRoundRobin defines the fallback load balancing strategy.
	DefaultBalanceRoundRobin = "roundRobin"
	// HeaderRequestSource is the metadata key for the originating service.
	HeaderRequestSource = "x-request-source"
	// HeaderRequestDest is the metadata key for the target service.
	HeaderRequestDest = "x-request-dest"
	// HeaderRequestApp is the metadata key for the application identifier.
	HeaderRequestApp = "x-request-app"
)

// SubConn wraps the gRPC SubConn with its resolved address metadata.
type SubConn struct {
	Address resolver.Address
	Conn    balancer.SubConn
}

// NewBalancerPicker is a factory function type that creates a new Picker
// based on the available sub-connections.
type NewBalancerPicker func(scs map[balancer.SubConn]base.SubConnInfo) Picker

// balancers maintains a global registry of available load balancing strategies.
var balancers = make(map[string]NewBalancerPicker)

// AddBalancer registers a custom load balancing strategy to the registry.
func AddBalancer(name string, newPicker NewBalancerPicker) {
	balancers[name] = newPicker
}

// BalanceBuilder implements the gRPC base.PickerBuilder interface.
// It acts as a bridge between gRPC's balancer framework and the custom pickers.
type BalanceBuilder struct {
	newPicker NewBalancerPicker
}

var _ base.PickerBuilder = &BalanceBuilder{}

// NewBalanceBuilder creates a gRPC balancer.Builder for the specified strategy name.
// If the requested strategy is not found, it falls back to the default Round Robin strategy.
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

// Build is called by gRPC when the connectivity state changes.
// It creates a new Picker instance using the ready sub-connections.
func (b *BalanceBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	scs := make(map[balancer.SubConn]base.SubConnInfo)
	for sc, sinfo := range info.ReadySCs {
		scs[sc] = sinfo
	}

	// NewPicker wraps the custom picker logic with the provided sub-connection map.
	return NewPicker(b.newPicker, scs)
}
