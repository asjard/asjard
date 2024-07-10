package grpc

import (
	"time"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
)

type ServerConfig struct {
	server.ServerConfig
	Options ServerOptionsConfig `json:"options"`
}

type ServerOptionsConfig struct {
	KeepaliveParams ServerKeepaliveParams `json:"keepaliveParams"`
}

type ServerKeepaliveParams struct {
	MaxConnectionIdle     utils.JSONDuration `json:"maxConnectionIdle"`
	MaxConnectionAge      utils.JSONDuration `json:"maxConnectionAge"`
	MaxConnectionAgeGrace utils.JSONDuration `json:"maxConnectionAgeGrace"`
	Time                  utils.JSONDuration `json:"time"`
	Timeout               utils.JSONDuration `json:"timeout"`
}

var defaultConfig ServerConfig = ServerConfig{
	Options: ServerOptionsConfig{
		KeepaliveParams: ServerKeepaliveParams{
			MaxConnectionIdle: utils.JSONDuration{Duration: time.Minute * 5},
			Time:              utils.JSONDuration{Duration: time.Second * 10},
			Timeout:           utils.JSONDuration{Duration: time.Second},
		},
	},
}
