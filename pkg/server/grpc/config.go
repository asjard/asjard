package grpc

import (
	"time"

	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
)

// Config represents the complete configuration for a gRPC server.
// It embeds the base server configuration and adds gRPC-specific options.
type Config struct {
	server.Config
	// Options contains specific gRPC protocol settings like keepalive.
	Options OptionsConfig `json:"options"`
}

// OptionsConfig wraps specific gRPC server parameters.
type OptionsConfig struct {
	// KeepaliveParams defines settings for server-side connection health monitoring.
	KeepaliveParams ServerKeepaliveParams `json:"keepaliveParams"`
}

// ServerKeepaliveParams maps to the standard grpc.KeepaliveParams.
// These settings help manage the lifecycle of a connection between a client and the server.
type ServerKeepaliveParams struct {
	// MaxConnectionIdle is a duration for the amount of time after which an idle connection
	// would be closed by sending a GoAway.
	MaxConnectionIdle utils.JSONDuration `json:"maxConnectionIdle"`

	// MaxConnectionAge is a duration for the maximum amount of time a connection may exist
	// before it will be closed by sending a GoAway.
	MaxConnectionAge utils.JSONDuration `json:"maxConnectionAge"`

	// MaxConnectionAgeGrace is an additive period after MaxConnectionAge after which
	// the connection will be forcibly closed.
	MaxConnectionAgeGrace utils.JSONDuration `json:"maxConnectionAgeGrace"`

	// Time is the duration after which a keepalive probe is sent if the server
	// sees no activity.
	Time utils.JSONDuration `json:"time"`

	// Timeout is the amount of time the server waits for a response to a keepalive
	// probe before closing the connection.
	Timeout utils.JSONDuration `json:"timeout"`
}

// defaultConfig initializes a gRPC configuration with sensible production defaults.
func defaultConfig() Config {
	return Config{
		// Inherit base settings for the gRPC protocol from the core server.
		Config: server.GetConfigWithProtocol(Protocol),
		Options: OptionsConfig{
			KeepaliveParams: ServerKeepaliveParams{
				// Close connections that have been idle for 5 minutes.
				MaxConnectionIdle: utils.JSONDuration{Duration: time.Minute * 5},
				// Send a ping every 10 seconds if no data is being sent.
				Time: utils.JSONDuration{Duration: time.Second * 10},
				// Wait 1 second for the ping response before considering the connection dead.
				Timeout: utils.JSONDuration{Duration: time.Second},
			},
		},
	}
}
