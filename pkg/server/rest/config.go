package rest

import (
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
)

const (
	defaultReadBufferSize  = 4096
	defaultWriteBufferSize = 4096
)

// ServerConfig 服务配置
type ServerConfig struct {
	server.ServerConfig
	Doc     ServerDocConfig     `json:"doc"`
	Openapi bool                `json:"openapi"`
	Options ServerOptionsConfig `json:"options"`
}

type ServerDocConfig struct {
	ErrPage string `json:"errPage"`
}

type ServerOptionsConfig struct {
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
	TCPKeepalive                       bool               `json:"tCPKeepalive"`
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

var defaultConfig ServerConfig = ServerConfig{
	ServerConfig: server.ServerConfig{
		Enabled: false,
	},
	Doc:     ServerDocConfig{},
	Options: ServerOptionsConfig{},
}
