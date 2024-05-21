package loadbalance

import "github.com/asjard/asjard/core/server"

// LoadBalanceRandom .
type LoadBalanceRandom struct{}

func init() {
	AddLoadbalancer(NewLoadBalanceRandom)
}

// NewLoadBalanceRandom .
func NewLoadBalanceRandom() (Loadbalancer, error) {
	return &LoadBalanceRandom{}, nil
}

// Pick .
func (l *LoadBalanceRandom) Pick([]*server.Instance) (*server.Instance, error) {
	return &server.Instance{}, nil
}

// Name .
func (l *LoadBalanceRandom) Name() string {
	return "random"
}
