/*
Package server 多协议服务维护
*/
package server

import (
	"fmt"
)

// Server 每个协议需要实现的内容
type Server interface {
	// 注册
	AddHandler(handler any) error
	// 服务启动
	Start(startErr chan error) error
	// 服务停止
	Stop()
	// 服务提供的协议
	Protocol() string
	// 服务监听地址列表
	ListenAddresses() AddressConfig
	// 是否已启用
	Enabled() bool
}

// NewServerFunc .
type NewServerFunc func(options *ServerOptions) (Server, error)

var (
	newServerFuncs = make(map[string]NewServerFunc)
)

// Init 服务初始化
// 初始化所有注册的服务
func Init() ([]Server, error) {
	// logger.Debug("start init server")
	// defer logger.Debug("init server done")
	var servers []Server
	for protocol, newServer := range newServerFuncs {
		interceptor, err := getChainUnaryInterceptors(protocol)
		if err != nil {
			return servers, err
		}
		server, err := newServer(&ServerOptions{
			Interceptor: interceptor,
		})
		if err != nil {
			return servers, err
		}
		// logger.Debug("server inited",
		// "protocol", server.Protocol())
		servers = append(servers, server)
	}
	return servers, nil
}

// AddServer 服务注册
func AddServer(protocol string, newServerFunc NewServerFunc) error {
	if _, ok := newServerFuncs[protocol]; ok {
		return fmt.Errorf("protocol %s server already exist", protocol)
	}
	newServerFuncs[protocol] = newServerFunc
	return nil
}
