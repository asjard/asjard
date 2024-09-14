package xasynq

import (
	"context"

	"github.com/hibiken/asynq"
)

// Context 存储asynq上下文
type Context struct {
	context.Context
	task *asynq.Task
}

// Payload 任务消息体
func (c Context) Payload() []byte {
	return c.task.Payload()
}

// Type 任务类型
func (c Context) Type() string {
	return c.task.Type()
}
