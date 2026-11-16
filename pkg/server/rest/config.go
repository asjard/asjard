package rest

import (
	"net/http"
	"time"

	"github.com/asjard/asjard/v2/core/config"
	"github.com/asjard/asjard/v2/core/server"
	"github.com/asjard/asjard/v2/utils"
)

const (
	defaultReadBufferSize  = 4096
	defaultWriteBufferSize = 4096
)

// Config represents the complete set of configuration options for the REST server.
type Config struct {
	server.Config               // Base server configuration (enabled, address, etc.)
	Routes        RoutesConfig  `json:"routes"`  // Custom route management settings
	Doc           DocConfig     `json:"doc"`     // Documentation and error page settings
	Openapi       OpenapiConfig `json:"openapi"` // OpenAPI/Swagger generation and UI settings
	Cors          CorsConfig    `json:"cors"`    // Cross-Origin Resource Sharing settings
	Options       OptionsConfig `json:"options"` // Low-level rest server tuning options
}

// RoutesConfig determines if route-related features are enabled.
type RoutesConfig struct {
	Enabled bool `json:"enabled"`
}

// DocConfig manages the high-level documentation links and error landing pages.
type DocConfig struct {
	ErrPage string `json:"errPage"`
}

// OpenapiConfig manages the automatic generation of OpenAPI specifications (YAML/JSON).
type OpenapiConfig struct {
	Enabled bool `json:"enabled"`
	// Page template for the Swagger UI (e.g., Swagger Petstore or Scalar).
	Page           string               `json:"page"`
	TermsOfService string               `json:"termsOfService"`
	License        OpenapiLicenseConfig `json:"license"`
	Scalar         ScalarOpenapiConfig  `json:"scalar"` // Configuration for Scalar API reference
	// Endpoint specifies the domain used in the OpenAPI spec; defaults to listenAddress.
	Endpoint string `json:"endpoint"`
}

// ScalarOpenapiConfig defines UI customization for the Scalar API documentation tool.
type ScalarOpenapiConfig struct {
	Theme              string            `json:"theme"`
	CDN                string            `json:"cdn"`
	SidebarVisibility  bool              `json:"sidebarVisibility"`
	HideModels         bool              `json:"hideModels"`
	HideDownloadButton bool              `json:"hideDownloadButton"`
	DarkMode           bool              `json:"darkMode"`
	HideClients        utils.JSONStrings `json:"hideClients"` // List of languages to hide in code snippets
	Authentication     string            `json:"authentication"`
}

type OpenapiLicenseConfig struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

// CorsConfig defines the security policy for browsers accessing the API from different origins.
type CorsConfig struct {
	allowAllOrigins bool // Internal flag for "*"
	// Disabled skips CORS handling entirely. Used by protocols such as MCP
	// that must not reject handshake requests with HTTP 403.
	Disabled         bool               `json:"disabled"`
	AllowOrigins     []string           `json:"allowOrigins"`
	AllowMethods     []string           `json:"allowMethods"`
	AllowHeaders     []string           `json:"allowHeaders"`
	ExposeHeaders    []string           `json:"exposeHeaders"`
	AllowCredentials bool               `json:"allowCredentials"`
	MaxAge           utils.JSONDuration `json:"maxAge"` // How long the browser caches the preflight response
}

// OptionsConfig contains low-level performance and timeout settings for the rest server.
type OptionsConfig struct {
	ReadTimeout        utils.JSONDuration `json:"readTimeout"`
	ReadHeaderTimeout  utils.JSONDuration `json:"readHeaderTimeout"`
	WriteTimeout       utils.JSONDuration `json:"writeTimeout"`
	IdleTimeout        utils.JSONDuration `json:"idleTimeout"`
	MaxHeaderBytes     int                `json:"maxHeaderBytes"`
	MaxRequestBodySize int                `json:"maxRequestBodySize"`
	DisableKeepalive   bool               `json:"disableKeepalive"`
}

// DefaultConfig initializes the REST server with safe, production-ready default values.
func DefaultConfig(protocol string) Config {
	return Config{
		Config: server.GetConfigWithProtocol(protocol),
		Doc: DocConfig{
			ErrPage: config.GetString("asjard.service.website", ""),
		},
		Openapi: OpenapiConfig{
			Page: "https://petstore.swagger.io/?url=%s/openapi.yml",
			License: OpenapiLicenseConfig{
				Name: "Apache 2.0",
				Url:  "http://www.apache.org/licenses/LICENSE-2.0.html",
			},
			Scalar: ScalarOpenapiConfig{
				Theme: "alternate",
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
			ReadTimeout:       utils.JSONDuration{Duration: 30 * time.Second},
			ReadHeaderTimeout: utils.JSONDuration{Duration: 10 * time.Second},
			IdleTimeout:       utils.JSONDuration{Duration: 60 * time.Second},
			// Default maximum request body size is 20MB.
			MaxRequestBodySize: 20 * 1024 * 1024,
		},
	}
}
