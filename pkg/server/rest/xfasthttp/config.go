package xfasthttp

import (
	"time"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
)

// Config represents the complete set of configuration options for the fasthttp REST server.
type Config struct {
	rest.Config
	Options OptionsConfig `json:"options"`
}

// OptionsConfig contains low-level performance and timeout settings for the fasthttp server.
type OptionsConfig struct {
	Concurrency                        int                `json:"concurrency"`
	ReadBufferSize                     int                `json:"readBufferSize"`
	WriteBufferSize                    int                `json:"writeBufferSize"`
	ReadTimeout                        utils.JSONDuration `json:"readTimeout"`
	WriteTimeout                       utils.JSONDuration `json:"writeTimeout"`
	IdleTimeout                        utils.JSONDuration `json:"idleTimeout"`
	MaxConnsPerIP                      int                `json:"maxConnsPerIP"`
	MaxRequestsPerConn                 int                `json:"maxRequestsPerConn"`
	MaxIdleWorkerDuration              utils.JSONDuration `json:"maxIdleWorkerDuration"`
	TCPKeepalivePeriod                 utils.JSONDuration `json:"tCPKeepalivePeriod"`
	MaxRequestBodySize                 int                `json:"maxRequestBodySize"`
	DisableKeepalive                   bool               `json:"disableKeepalive"`
	TCPKeepalive                       bool               `json:"tcpKeepalive"`
	ReduceMemoryUsage                  bool               `json:"reduceMemoryUsage"`
	GetOnly                            bool               `json:"getOnly"`
	DisablePreParseMultipartForm       bool               `json:"disablePreParseMultipartForm"`
	LogAllErrors                       bool               `json:"logAllErrors"`
	SecureErrorLogMessage              bool               `json:"secureErrorLogMessage"`
	DisableHeaderNamesNormalizing      bool               `json:"disableHeaderNamesNormalizing"`
	SleepWhenConcurrencyLimitsExceeded utils.JSONDuration `json:"sleepWhenConcurrencyLimitsExceeded"`
	NoDefaultServerHeader              bool               `json:"noDefaultServerHeader"`
	NoDefaultDate                      bool               `json:"noDefaultDate"`
	NoDefaultContentType               bool               `json:"noDefaultContentType"`
	KeepHijackedConns                  bool               `json:"keepHijackedConns"`
	CloseOnShutdown                    bool               `json:"closeOnShutdown"`
	StreamRequestBody                  bool               `json:"streamRequestBody"`
}

// DefaultConfig initializes the fasthttp REST server with safe, production-ready default values.
func DefaultConfig(protocol string) Config {
	return Config{
		Config: rest.DefaultConfig(protocol),
		Options: OptionsConfig{
			ReadTimeout:        utils.JSONDuration{Duration: 30 * time.Second},
			IdleTimeout:        utils.JSONDuration{Duration: 60 * time.Second},
			MaxRequestBodySize: 20 * 1024 * 1024,
		},
	}
}
