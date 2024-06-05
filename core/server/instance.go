package server

import (
	"errors"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
)

// Instance 服务实例详情
type Instance struct {
	// 所属项目
	App string
	// 所属环境
	Environment string
	// 所属区域
	Region string
	// 可用区
	AZ string

	// 服务ID
	ID string
	// 服务名称
	Name string
	// 服务版本
	Version string
	// 服务端口列表
	// key为协议名称
	// value-key 监听地址名称
	// value-value 监听地址列表
	Endpoints map[string]map[string][]string
	// 服务元数据
	MetaData map[string]string
}

// ServiceInstance 服务实例
var serviceInstance *Instance
var sonce sync.Once

// GetInstance 返回服务实例详情
func GetInstance() *Instance {
	sonce.Do(func() {
		serviceInstance = NewInstance()
	})
	return serviceInstance
}

// NewInstance .
func NewInstance() *Instance {
	metadata := make(map[string]string)
	if err := config.GetWithUnmarshal("instance.metadata", &metadata); err != nil {
		logger.Errorf("get instance fail[%s]", err.Error())
	}
	return &Instance{
		App:         runtime.APP,
		Environment: runtime.Environment,
		Region:      runtime.Region,
		ID:          runtime.ServiceID,
		Name:        runtime.Name,
		Version:     runtime.Version,
		Endpoints:   make(map[string]map[string][]string),
		MetaData:    metadata,
	}
}

// AddEndpoint 添加服务端口
func (s *Instance) AddEndpoint(protocol string, endpoints map[string]string) error {
	if len(endpoints) == 0 {
		return nil
	}
	if protocol == "" {
		return errors.New("endpoint protocol is must")
	}
	for name, address := range endpoints {
		listenAddress, err := utils.GetListenAddress(address)
		if err != nil {
			return err
		}
		if _, ok := s.Endpoints[protocol]; ok {
			s.Endpoints[protocol][name] = append(s.Endpoints[protocol][name], listenAddress)
		} else {
			s.Endpoints[protocol] = map[string][]string{
				name: {listenAddress},
			}
		}
	}
	return nil
}

// AddEndpoints 添加服务端口
func (s *Instance) AddEndpoints(protocol string, endpoints map[string][]string) error {
	if len(endpoints) == 0 {
		return nil
	}
	if protocol == "" {
		return errors.New("endpoint protocol is must")
	}
	for name, addresses := range endpoints {
		for _, address := range addresses {
			listenAddress, err := utils.GetListenAddress(address)
			if err != nil {
				return err
			}
			if _, ok := s.Endpoints[protocol]; ok {
				s.Endpoints[protocol][name] = append(s.Endpoints[protocol][name], listenAddress)
			} else {
				s.Endpoints[protocol] = map[string][]string{
					name: {listenAddress},
				}
			}
		}
	}
	return nil
}

// SetMetadata 设置元数据
func (s *Instance) SetMetadata(key, value string) {
	s.MetaData[key] = value
}
