package server

import (
	"github.com/asjard/asjard/core/logger"
)

// Server 每个协议需要实现的内容
type Server interface {
	// 注册
	AddHandler(handler any) error
	// 请求处理前
	// BeforeHandle()
	// 请求处理
	// Handle(request *Request) (*Response, error)
	// 请求前
	// PreRequest()
	// 请求后
	// AfterRequest()
	// 服务启动
	Start() error
	// 服务停止
	Stop()
	// 服务提供的协议
	Protocol() string
	// 服务监听地址列表
	ListenAddresses() []*EndpointAddress
}

// Handler .
type Handler interface {
	Protocol() string
}

// NewServerFunc .
type NewServerFunc func() (Server, error)

var (
	newServerFuncs []NewServerFunc
)

// Init 服务初始化
// 初始化所有注册的服务
func Init() ([]Server, error) {
	logger.Debug("start init server")
	defer logger.Debug("init server done")
	var servers []Server
	for _, newServer := range newServerFuncs {
		server, err := newServer()
		if err != nil {
			return servers, err
		}
		logger.Debugf("server '%s' inited", server.Protocol())
		servers = append(servers, server)
	}
	return servers, nil
}

// AddServer 服务注册
func AddServer(newServerFunc NewServerFunc) error {
	newServerFuncs = append(newServerFuncs, newServerFunc)
	return nil
}
