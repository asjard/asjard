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

// Initiator 初始化需要实现的方法
type Initiator interface {
	// 启动
	Start() error
	// 停止
	Stop()
}

var (
	bootstrapHandlers []Initiator
	bootstrapedMap    = make(map[Initiator]struct{})

	initiatorHandlers []Initiator
	initiatorMap      = make(map[Initiator]struct{})
)

// AddBootstrap 添加启动方法
// 初始化后，服务启动前执行
func AddBootstrap(handler Initiator) {
	if _, ok := bootstrapedMap[handler]; !ok {
		bootstrapHandlers = append(bootstrapHandlers, handler)
		bootstrapedMap[handler] = struct{}{}
	}
}

// AddBootstraps 批量添加启动方法
func AddBootstraps(handlers ...Initiator) {
	for _, handler := range handlers {
		AddBootstrap(handler)
	}
}

// AddInitator 添加初始化方法
// 加载到env,file环境变量后执行
func AddInitiator(handler Initiator) {
	if _, ok := initiatorMap[handler]; !ok {
		initiatorHandlers = append(initiatorHandlers, handler)
		initiatorMap[handler] = struct{}{}
	}
}

// AddInitiators 批量添加初始化方法
func AddInitiators(handlers ...Initiator) {
	for _, handler := range handlers {
		AddInitiator(handler)
	}
}

// Init 初始化
func Init() error {
	for _, handler := range initiatorHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Bootstrap 系统启动
func Bootstrap() error {
	for _, handler := range bootstrapHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Shutdown 系统停止
func Shutdown() {
	for idx := len(bootstrapHandlers) - 1; idx >= 0; idx-- {
		bootstrapHandlers[idx].Stop()
	}
	for idx := len(initiatorHandlers) - 1; idx >= 0; idx-- {
		initiatorHandlers[idx].Stop()
	}
}
