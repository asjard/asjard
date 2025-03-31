package server

import (
	"errors"
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
)

// Service 服务详情
type Service struct {
	runtime.APP
	// 服务端口列表
	// key为协议名称
	// value-key 监听地址名称
	// value-value 监听地址列表，有可能有多个实例列表所以是个列表
	Endpoints map[string]*Endpoint
	em        sync.RWMutex
}

// ServiceInstance 服务实例
var service *Service
var sonce sync.Once

// GetInstance 返回服务实例详情
func GetService() *Service {
	sonce.Do(func() {
		service = NewService()
	})
	return service
}

// NewInstance .
func NewService() *Service {
	return &Service{
		APP:       runtime.GetAPP(),
		Endpoints: make(map[string]*Endpoint),
	}
}

// AddEndpoint 添加服务端口
func (s *Service) AddEndpoint(protocol string, address AddressConfig) error {
	if protocol == "" {
		return errors.New("endpoint protocol is must")
	}
	s.em.RLock()
	if _, ok := s.Endpoints[protocol]; !ok {
		s.Endpoints[protocol] = &Endpoint{}
	}
	s.em.RUnlock()

	s.em.Lock()
	if address.Listen != "" {
		listenAddress, err := utils.GetListenAddress(address.Listen)
		if err != nil {
			return err
		}
		logger.Debug("service listen address", "protocol", protocol, "listen", listenAddress)
		s.Endpoints[protocol].Listen = append(s.Endpoints[protocol].Listen, listenAddress)
	}
	if address.Advertise != "" {
		s.Endpoints[protocol].Advertise = append(s.Endpoints[protocol].Advertise, address.Advertise)
	}
	s.em.Unlock()
	return nil
}

// GetListenAddresses 获取监听地址
func (s *Service) GetListenAddresses(protocol string) []string {
	s.em.RLock()
	endpoint, ok := s.Endpoints[protocol]
	s.em.RUnlock()
	if !ok {
		return []string{}
	}
	return endpoint.Listen
}

// GetAdvertiseAddresses 获取广播地址
func (s *Service) GetAdvertiseAddresses(protocol string) []string {
	s.em.RLock()
	endpoint, ok := s.Endpoints[protocol]
	s.em.RUnlock()
	if !ok {
		return []string{}
	}
	return endpoint.Advertise
}

func (s *Service) GetEndpoint(protocol string) (*Endpoint, bool) {
	s.em.RLock()
	endpoint, ok := s.Endpoints[protocol]
	s.em.RUnlock()
	return endpoint, ok
}
