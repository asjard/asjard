package xasynq

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/hibiken/asynq"
)

// Handler is the interface that specific task services must implement.
// It allows the server to discover task types and their associated processing logic.
type Handler interface {
	// AsynqServiceDesc returns the service metadata, mapping task type strings
	// to their respective execution functions.
	AsynqServiceDesc() *ServiceDesc
}

// GlobaltHandler defines the hooks for managing the server's runtime behavior.
// These methods control how the server reacts to failures, manages task groups,
// and monitors system health globally.
type GlobaltHandler interface {
	// BaseContext provides the root context for all processed tasks.
	BaseContext() func() context.Context
	// RetryDelayFunc calculates how long to wait before retrying a failed task.
	RetryDelayFunc() func(n int, e error, t *asynq.Task) time.Duration
	// IsFailure determines if a returned error should count as a task failure.
	IsFailure() func(error) bool
	// HealthCheckFunc is called periodically to report the worker's status.
	HealthCheckFunc() func(error)
	// ErrorHandler is triggered when a task exceeds its retry limit or fails.
	ErrorHandler() asynq.ErrorHandler
	// GroupAggregator handles logic for merging multiple tasks into a group.
	GroupAggregator() asynq.GroupAggregator
}

// defaultGlobalHandler provides a baseline implementation of the GlobaltHandler.
type defaultGlobalHandler struct{}

// WithGlobalHanler allows developers to override the default system behavior
// with a custom global controller.
func WithGlobalHanler(handler GlobaltHandler) {
	globalHandler = handler
}

// BaseContext returns a standard background context by default.
func (h defaultGlobalHandler) BaseContext() func() context.Context {
	return func() context.Context {
		return context.Background()
	}
}

// RetryDelayFunc returns nil by default, letting Asynq use its internal
// exponential backoff strategy.
func (h defaultGlobalHandler) RetryDelayFunc() func(n int, e error, t *asynq.Task) time.Duration {
	return nil
}

// IsFailure treats any non-nil error as a task failure.
func (h defaultGlobalHandler) IsFailure() func(err error) bool {
	return func(err error) bool {
		return err != nil
	}
}

// HealthCheckFunc logs an error message if the server's health check fails.
func (h defaultGlobalHandler) HealthCheckFunc() func(err error) {
	return func(err error) {
		if err != nil {
			logger.Error("asynq health check fail", "err", err)
		}
	}
}

// ErrorHandler returns the current instance as it implements the asynq.ErrorHandler interface.
func (h defaultGlobalHandler) ErrorHandler() asynq.ErrorHandler {
	return h
}

// GroupAggregator is disabled by default.
func (h defaultGlobalHandler) GroupAggregator() asynq.GroupAggregator {
	return nil
}

// HandleError is the final destination for task failures, logging the task
// type and payload to assist in debugging.
func (h defaultGlobalHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	logger.Error("asynq handle error", "type", task.Type(), "payload", task.Payload(), "err", err)
}
