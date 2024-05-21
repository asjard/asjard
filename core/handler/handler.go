package handler

import (
	"github.com/asjard/asjard/core/handler/loadbalance"
	"github.com/asjard/asjard/core/logger"
)

// Handler 请求处理
type Handler interface {
}

// Init handler初始化
func Init() error {
	logger.Debug("Start init handler")
	defer logger.Debug("init handler Done")
	if err := loadbalance.Init(); err != nil {
		return err
	}
	return nil
}
