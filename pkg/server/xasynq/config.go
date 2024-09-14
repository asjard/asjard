package xasynq

import (
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
)

type Config struct {
	server.Config
	// redis名称asjard.stores.redis.clients.{名称}
	Redis   string  `json:"redis"`
	Options Options `json:"options"`
}

type Options struct {
	Concurrency              int                `json:"concurrency"`
	Queue                    map[string]int     `json:"queue"`
	StrictPriority           bool               `json:"strictPriority"`
	ShutdownTimeout          utils.JSONDuration `json:"shutdownDuration"`
	HealthCheckInterval      utils.JSONDuration `json:"healthCheckInterval"`
	DelayedTaskCheckInterval utils.JSONDuration `json:"delayedTaskCheckInterval"`
	GroupGracePeriod         utils.JSONDuration `json:"groupGracePeriod"`
	GroupMaxDelay            utils.JSONDuration `json:"groupMaxDelay"`
	GroupMaxSize             int                `json:"groupMaxSize"`
}

func defaultConfig() Config {
	return Config{}
}
