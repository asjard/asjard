package loadbalance

import (
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// Loadbalancer 负载均衡器需要实现的方法
type Loadbalancer interface {
	// 服务选择
	Pick([]*server.Instance) (*server.Instance, error)
	// 负载均衡器名称
	Name() string
}

// NewLoadbalancerFunc .
type NewLoadbalancerFunc func() (Loadbalancer, error)

// Loadbalance .
type Loadbalance struct {
	loadbalancers map[string]Loadbalancer
	config        *Config
	sync.RWMutex
}

var newLoadbalanceFuncs []NewLoadbalancerFunc
var lbManager *Loadbalance

func init() {
	lbManager = &Loadbalance{
		loadbalancers: make(map[string]Loadbalancer),
	}
}

// AddLoadbalancer .
func AddLoadbalancer(newFunc NewLoadbalancerFunc) {
	newLoadbalanceFuncs = append(newLoadbalanceFuncs, newFunc)
}

// Init 负载均衡初始化
func Init() error {
	logger.Debug("start init loadbalance")
	defer logger.Debug("init loadbalance Done")
	lbManager.Lock()
	defer lbManager.Unlock()
	for _, newLoadbalanceFunc := range newLoadbalanceFuncs {
		newLoadbalance, err := newLoadbalanceFunc()
		if err != nil {
			return err
		}
		lbManager.loadbalancers[newLoadbalance.Name()] = newLoadbalance
	}
	lbManager.config = loadConfig()
	return nil
}
