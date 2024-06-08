package bootstrap

import "github.com/asjard/asjard/core/logger"

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
	logger.Debug("bootstrap Start")
	defer logger.Debug("bootstrap Done")
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
