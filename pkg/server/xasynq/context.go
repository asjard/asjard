package xasynq

import (
	"context"

	"github.com/hibiken/asynq"
)

// Context wraps the standard Go context and the Asynq task object.
// It acts as the primary data carrier passed into task handlers, providing
// both execution control (deadlines, cancellations) and task data.
type Context struct {
	context.Context
	// task represents the unit of work fetched from Redis,
	// containing the unique identifier, type, and payload.
	task *asynq.Task
}

// Payload returns the raw byte data associated with the task.
// Handlers typically unmarshal this data (e.g., from JSON or Protobuf)
// to retrieve the parameters needed to execute the job.
func (c Context) Payload() []byte {
	return c.task.Payload()
}

// Type returns the string identifier for the task.
// This is used by the Asynq multiplexer to route the task to the
// correct handler function (e.g., "email:send" or "report:generate").
func (c Context) Type() string {
	return c.task.Type()
}
