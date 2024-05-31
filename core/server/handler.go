package server

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
)

// ServerHandlers .
type ServerHandlers struct {
	// 请求前中间件
	BeforeRequestHandlers []ServerHandler
	// 请求后中间件
	AfterRequestHandlers []ServerHandler
}

// ServerHandler 中间件需要实现的方法
type ServerHandler interface {
	Name() string
	Handle(ctx *Context)
}

var serverHandlers []ServerHandler

// AddServerHandler 添加中间件
func AddServerHandler(middlewarer ServerHandler) {
	serverHandlers = append(serverHandlers, middlewarer)
}

// BeforeRequestHandle 请求前处理
func (m *ServerHandlers) BeforeRequestHandle(ctx *Context) {
	for _, mdw := range m.BeforeRequestHandlers {
		mdw.Handle(ctx)
	}
}

// AfterRequestHandle 请求后处理
func (m *ServerHandlers) AfterRequestHandle(ctx *Context) {
	for _, mdw := range m.AfterRequestHandlers {
		mdw.Handle(ctx)
	}
}

// GetHandlersByProtocol 根据协议获取中间件
// 先获取协议的，协议没有的话获取全局的
func getHandlersByProtocol(protocol string) *ServerHandlers {
	return &ServerHandlers{
		BeforeRequestHandlers: getMiddlewareByProtocolPosition(protocol, "beforeRequest"),
		AfterRequestHandlers:  getMiddlewareByProtocolPosition(protocol, "afterRequest"),
	}
}

func getMiddlewareByProtocolPosition(protocol, position string) []ServerHandler {
	var middlewarers []ServerHandler
	for _, middlewareName := range config.GetStrings(fmt.Sprintf("servers.%s.middlewares.%s", protocol, position),
		config.GetStrings(fmt.Sprintf("servers.middlewares.%s", position),
			[]string{},
			config.WithDelimiter(constant.DefaultDelimiter)),
		config.WithDelimiter(constant.DefaultDelimiter)) {
		for _, smdw := range serverHandlers {
			if smdw.Name() == middlewareName {
				middlewarers = append(middlewarers, smdw)
			}
		}
	}
	return middlewarers
}
