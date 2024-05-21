package server

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
)

// Service .
type Service struct {
	// 服务名称
	Name string
	// 协议
	Protocol    string
	LoadBalance string
}

// NewService 用以客户端访问
// 服务名称和协议
func NewService(name, protocol string) *Service {
	return &Service{
		Name:     name,
		Protocol: protocol,
		LoadBalance: config.GetString(fmt.Sprintf("loadbalance.services.%s.strategy", name),
			config.GetString("loadbalance.strategy", "")),
	}
}
