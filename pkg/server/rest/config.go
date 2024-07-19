package rest

import (
	"net/http"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
)

const (
	defaultReadBufferSize  = 4096
	defaultWriteBufferSize = 4096
)

// Config 服务配置
type Config struct {
	server.Config
	Doc     DocConfig     `json:"doc"`
	Openapi OpenapiConfig `json:"openapi"`
	Cors    CorsConfig    `json:"cors"`
	Options OptionsConfig `json:"options"`
}

type DocConfig struct {
	ErrPage string `json:"errPage"`
}

type OpenapiConfig struct {
	Enabled bool `json:"enabled"`
	// https://petstore.swagger.io/?url=http://%s/openapi.yml
	// https://petstore.swagger.io/?url=http://127.0.0.1:7030/openapi.yml
	// https://authress-engineering.github.io/openapi-explorer/?url=http://%s/openapi.yml
	// https://authress-engineering.github.io/openapi-explorer/?url=http://127.0.0.1:7030/openapi.yml
	Page           string               `json:"page"`
	TermsOfService string               `json:"termsOfService"`
	License        OpenapiLicenseConfig `json:"license"`
}

type OpenapiLicenseConfig struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type CorsConfig struct {
	allowAllOrigins  bool
	AllowOrigins     []string           `json:"allowOrigins"`
	AllowMethods     []string           `json:"allowMethods"`
	AllowHeaders     []string           `json:"allowHeaders"`
	ExposeHeaders    []string           `json:"exposeHeaders"`
	AllowCredentials bool               `json:"allowCredentials"`
	MaxAge           utils.JSONDuration `json:"maxAge"`
}

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

func defaultConfig() Config {
	return Config{
		Config: server.GetConfigWithProtocol(Protocol),
		Doc: DocConfig{
			ErrPage: config.GetString(constant.ConfigWebsite, ""),
		},
		Openapi: OpenapiConfig{
			Page: "https://petstore.swagger.io/?url=http://%s/openapi.yml",
			License: OpenapiLicenseConfig{
				Name: "Apache 2.0",
				Url:  "http://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		Cors: CorsConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type"},
			AllowCredentials: false,
			MaxAge:           utils.JSONDuration{Duration: 12 * time.Hour},
		},
		Options: OptionsConfig{
			MaxRequestBodySize: 10 * 1024 * 1024,
		},
	}
}
