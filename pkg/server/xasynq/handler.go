package xasynq

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/hibiken/asynq"
)

// Handler 服务需要实现的方法
type Handler interface {
	AsynqServiceDesc() *ServiceDesc
}

// GlobaltHandler 全局处理器
type GlobaltHandler interface {
	BaseContext() func() context.Context
	RetryDelayFunc() func(n int, e error, t *asynq.Task) time.Duration
	IsFailure() func(error) bool
	HealthCheckFunc() func(error)
	ErrorHandler() asynq.ErrorHandler
	GroupAggregator() asynq.GroupAggregator
}

type defaultGlobalHandler struct{}

// WithGlobalHanler 设置全局处理器
func WithGlobalHanler(handler GlobaltHandler) {
	globalHandler = handler
}

func (h defaultGlobalHandler) BaseContext() func() context.Context {
	return func() context.Context {
		return context.Background()
	}
}

func (h defaultGlobalHandler) RetryDelayFunc() func(n int, e error, t *asynq.Task) time.Duration {
	return nil
}

func (h defaultGlobalHandler) IsFailure() func(err error) bool {
	return func(err error) bool {
		return err != nil
	}
}

func (h defaultGlobalHandler) HealthCheckFunc() func(err error) {
	return func(err error) {
		if err != nil {
			logger.Error("asynq health check fail", "err", err)
		}
	}
}

func (h defaultGlobalHandler) ErrorHandler() asynq.ErrorHandler {
	return h
}

func (h defaultGlobalHandler) GroupAggregator() asynq.GroupAggregator {
	return nil
}

func (h defaultGlobalHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	logger.Error("asynq handle error", "type", task.Type(), "payload", task.Payload(), "err", err)
}
