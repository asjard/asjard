package handlers

import (
	"github.com/asjard/asjard/core/server"
)

type serverHandler struct {
	name    string
	handler any
}

var serverDefaultHandlers []*serverHandler

// AddServerDefaultHandler 添加服务默认handler
func AddServerDefaultHandler(name string, handler any) {
	serverDefaultHandlers = append(serverDefaultHandlers, &serverHandler{
		name:    name,
		handler: handler,
	})
}

// GetServerDefaultHandlers 获取服务默认handler列表
func GetServerDefaultHandlers(protocol string) []any {
	var handlers []any
	conf := server.GetConfigWithProtocol(protocol)
	for _, name := range conf.DefaultHandlers {
		for _, defaultHandler := range serverDefaultHandlers {
			if name == defaultHandler.name {
				handlers = append(handlers, defaultHandler.handler)
			}
		}
	}
	return handlers
}
