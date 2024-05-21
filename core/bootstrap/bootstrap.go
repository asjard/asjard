package bootstrap

import "github.com/asjard/asjard/core/logger"

// BootstrapHandler 启动引导需实现的方法
type BootstrapHandler interface {
	Start() error
	Stop()
}

var bootstrapHandlers []BootstrapHandler

// AddBootstrap 添加启动方法
func AddBootstrap(handler BootstrapHandler) {
	bootstrapHandlers = append(bootstrapHandlers, handler)
}

// Start 系统启动
func Start() error {
	logger.Debug("bootstrap Start")
	defer logger.Debug("bootstrap Done")
	for _, handler := range bootstrapHandlers {
		if err := handler.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop 系统停止
func Stop() {
	for _, handler := range bootstrapHandlers {
		handler.Stop()
	}
}
