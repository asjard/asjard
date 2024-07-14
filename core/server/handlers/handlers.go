package handlers

import (
	"sync"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/server"
)

type serverHandler struct {
	name    string
	handler any
}

var (
	serverDefaultHandlers = make(map[string][]*serverHandler)
	sdhm                  sync.RWMutex
)

// AddServerDefaultHandler 添加服务默认handler
func AddServerDefaultHandler(name string, handler any, protocols ...string) {
	if len(protocols) == 0 {
		protocols = []string{constant.AllProtocol}
	}
	sdhm.Lock()
	defaultHandler := &serverHandler{
		name:    name,
		handler: handler,
	}
	for _, protocol := range protocols {
		if _, ok := serverDefaultHandlers[protocol]; ok {
			serverDefaultHandlers[protocol] = append(serverDefaultHandlers[protocol], defaultHandler)
		} else {
			serverDefaultHandlers[protocol] = []*serverHandler{defaultHandler}
		}
	}
	sdhm.Unlock()
}

// GetServerDefaultHandlers 获取服务默认handler列表
func GetServerDefaultHandlers(protocol string) []any {
	var handlers []any
	conf := server.GetConfigWithProtocol(protocol)
	sdhm.RLock()
	defaultHandlers := serverDefaultHandlers[protocol]
	defaultHandlers = append(defaultHandlers, serverDefaultHandlers[constant.AllProtocol]...)
	sdhm.RUnlock()
	for _, name := range conf.DefaultHandlers {
		for _, defaultHandler := range defaultHandlers {
			if name == defaultHandler.name {
				handlers = append(handlers, defaultHandler.handler)
			}
		}
	}
	return handlers
}
