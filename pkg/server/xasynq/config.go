package xasynq

import (
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
)

// Config represents the complete configuration for the Asynq server.
// It integrates standard server settings with Asynq-specific Redis and worker options.
type Config struct {
	server.Config

	// Redis identifies which Redis client to use from the central store.
	// Maps to: asjard.stores.redis.clients.{Redis}
	Redis string `json:"redis"`

	// Options contains the performance and behavioral settings for the Asynq processor.
	Options Options `json:"options"`
}

// Options contains the tuning parameters for the underlying Asynq server.
type Options struct {
	// Concurrency specifies the maximum number of concurrent workers to process tasks.
	Concurrency int `json:"concurrency"`

	// Queue maps queue names to their priority levels (e.g., {"critical": 6, "default": 3, "low": 1}).
	Queue map[string]int `json:"queue"`

	// StrictPriority, if true, ensures workers always process all tasks from a higher-priority
	// queue before moving to a lower-priority one.
	StrictPriority bool `json:"strictPriority"`

	// ShutdownTimeout determines how long to wait for active tasks to finish during a graceful shutdown.
	ShutdownTimeout utils.JSONDuration `json:"shutdownDuration"`

	// HealthCheckInterval defines the frequency of background health checks for the worker.
	HealthCheckInterval utils.JSONDuration `json:"healthCheckInterval"`

	// DelayedTaskCheckInterval defines how often to check for scheduled/delayed tasks that are ready to run.
	DelayedTaskCheckInterval utils.JSONDuration `json:"delayedTaskCheckInterval"`

	// GroupGracePeriod defines the time to wait for more tasks to join a group before processing.
	GroupGracePeriod utils.JSONDuration `json:"groupGracePeriod"`

	// GroupMaxDelay is the maximum time a task can wait in a group before being processed.
	GroupMaxDelay utils.JSONDuration `json:"groupMaxDelay"`

	// GroupMaxSize is the maximum number of tasks that can be aggregated into a single group.
	GroupMaxSize int `json:"groupMaxSize"`
}

// defaultConfig returns an empty Config instance.
// Default values are typically populated by the Asynq library or the config loader.
func defaultConfig() Config {
	return Config{}
}
