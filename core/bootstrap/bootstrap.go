/*
Package bootstrap 服务初始化后启动之前执行的一些初始化任务，加载一些内建功能
*/
package bootstrap

import (
	// 加载加解密组件
	_ "github.com/asjard/asjard/pkg/security"
	// 服务端拦截器
	_ "github.com/asjard/asjard/pkg/server/interceptors"
	// 默认handler
	_ "github.com/asjard/asjard/pkg/server/handlers"
	// 客户端拦截器
	_ "github.com/asjard/asjard/pkg/client/interceptors"
	// 导入pprof包, 这样就不需要在main函数中导入了
	_ "github.com/asjard/asjard/pkg/server/pprof"
	// 导入内存配置源
	_ "github.com/asjard/asjard/pkg/config/mem"
	// 导入环境变量配置源
	_ "github.com/asjard/asjard/pkg/config/env"
)

// BootstrapHandler 启动引导需实现的方法
type BootstrapHandler interface {
	// 启动时执行
	Bootstrap() error
	// 停止时执行
	Shutdown()
}

var bootstrapHandlers []BootstrapHandler

var bootstrapedMap = make(map[BootstrapHandler]struct{})

// AddBootstrap 添加启动方法
func AddBootstrap(handler BootstrapHandler) {
	if _, ok := bootstrapedMap[handler]; !ok {
		bootstrapHandlers = append(bootstrapHandlers, handler)
		bootstrapedMap[handler] = struct{}{}
	}
}

// Start 系统启动
func Start() error {
	for _, handler := range bootstrapHandlers {
		if err := handler.Bootstrap(); err != nil {
			return err
		}
	}
	return nil
}

// Stop 系统停止
func Stop() {
	for _, handler := range bootstrapHandlers {
		handler.Shutdown()
	}
}
