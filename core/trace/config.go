package trace

import (
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/utils"
)

// Config holds the settings for the distributed tracing exporter.
type Config struct {
	// Enabled determines if tracing data should be collected and exported.
	Enabled bool `json:"enabled"`

	// Endpoint is the target address for the trace collector.
	// It must include the protocol scheme:
	// Example: http://127.0.0.1:4318 (OTLP/HTTP)
	// Example: grpc://127.0.0.1:4317 (OTLP/gRPC)
	Endpoint string `json:"endpoint"`

	// Timeout defines the maximum time allowed for an export request to complete.
	Timeout utils.JSONDuration `json:"timeout"`

	// CertFile is the relative path to the client certificate for TLS.
	CertFile string `json:"certFile"`

	// KeyFile is the relative path to the client private key for TLS.
	KeyFile string `json:"keyFile"`

	// CaFile is the relative path to the Certificate Authority file to verify the collector.
	CaFile string `json:"cafile"`
}

// defaultTraceConfig sets the baseline values for tracing,
// ensuring a 1-second timeout if not explicitly configured.
var defaultTraceConfig = Config{
	Timeout: utils.JSONDuration{Duration: time.Second},
}

// GetConfig retrieves the trace configuration from the global config system.
// It looks for settings under the "asjard.trace" prefix.
func GetConfig() *Config {
	conf := defaultTraceConfig
	// Unmarshal config data from files/env into the Config struct.
	config.GetWithUnmarshal("asjard.trace", &conf)
	return &conf
}
