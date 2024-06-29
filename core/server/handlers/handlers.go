package handlers

import (
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
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
	for _, name := range config.GetStrings(fmt.Sprintf(constant.ConfigServerDefaultHandlersWithProtocol, protocol),
		config.GetStrings(constant.ConfigServerDefaultHandlers, []string{"health"})) {
		for _, defaultHandler := range serverDefaultHandlers {
			if name == defaultHandler.name {
				handlers = append(handlers, defaultHandler.handler)
			}
		}
	}
	return handlers
}
