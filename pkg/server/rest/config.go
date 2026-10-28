package rest

import (
	"net/http"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/utils"
)

// Config represents the protocol-level REST/HTTP server options shared by engines.
type Config struct {
	server.Config               // Base server configuration (enabled, address, etc.)
	Routes        RoutesConfig  `json:"routes"`  // Custom route management settings
	Doc           DocConfig     `json:"doc"`     // Documentation and error page settings
	Openapi       OpenapiConfig `json:"openapi"` // OpenAPI/Swagger generation and UI settings
	Cors          CorsConfig    `json:"cors"`    // Cross-Origin Resource Sharing settings
	I18n          I18nConfig    `json:"i18n"`    // Internationalization settings
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
	AllowOrigins     []string           `json:"allowOrigins"`
	AllowMethods     []string           `json:"allowMethods"`
	AllowHeaders     []string           `json:"allowHeaders"`
	ExposeHeaders    []string           `json:"exposeHeaders"`
	AllowCredentials bool               `json:"allowCredentials"`
	MaxAge           utils.JSONDuration `json:"maxAge"` // How long the browser caches the preflight response
}

// I18nConfig manages whether internationalized error responses are enabled.
type I18nConfig struct {
	Enabled bool `json:"enabled"`
}

// DefaultConfig initializes shared REST/HTTP protocol defaults.
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
	}
}
