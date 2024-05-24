package server

import (
	"errors"
	"fmt"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
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
	Endpoints []*Endpoint
	// 服务元数据
	MetaData map[string]string
}

// ServiceInstance 服务实例
var serviceInstance *Instance
var sonce sync.Once

// GetInstance 返回服务实例详情
func GetInstance() *Instance {
	sonce.Do(func() {
		metadata := make(map[string]string)
		if err := config.GetWithUnmarshal("instance.metadata", &metadata); err != nil {
			logger.Errorf("get instance fail[%s]", err.Error())
		}
		serviceInstance = &Instance{
			App:         runtime.APP,
			Environment: runtime.Environment,
			Region:      runtime.Region,
			ID:          runtime.ServiceID,
			Name:        runtime.Name,
			Version:     runtime.Version,
			MetaData:    metadata,
		}
	})
	return serviceInstance
}

// AddEndpoint 添加服务端口
func (s *Instance) AddEndpoint(endpoint *Endpoint) error {
	if endpoint == nil {
		return nil
	}
	if endpoint.Protocol == "" {
		return errors.New("endpoint protocol is must")
	}
	for _, ed := range s.Endpoints {
		if ed.Protocol == endpoint.Protocol {
			return fmt.Errorf("protocol '%s' already exist", endpoint.Protocol)
		}
	}
	s.Endpoints = append(s.Endpoints, endpoint)
	return nil
}
