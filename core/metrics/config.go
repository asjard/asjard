package metrics

import (
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/utils"
)

const (
	// AllCollectors is a wildcard constant used to enable every available metric collector.
	AllCollectors = "*"
)

// Config represents the monitoring and metrics settings for the application.
type Config struct {
	// Enabled determines if the metrics collection system is active.
	Enabled bool `json:"enabled"`

	// allCollectors is an internal flag used to check if the wildcard "*" was provided.
	allCollectors bool

	// Collectors defines a list of custom or third-party metric collectors to enable.
	Collectors utils.JSONStrings `json:"collectors"`

	// BuiltInCollectors defines the standard framework metrics enabled by default
	// (e.g., Go runtime stats, process info).
	BuiltInCollectors utils.JSONStrings `json:"builtInCollectors"`

	// PushGateway configures how metrics are "pushed" to a Prometheus PushGateway.
	// This is useful for short-lived jobs or environments where Prometheus cannot scrape the app.
	PushGateway PushGatewayConfig `json:"pushGateway"`
}

// PushGatewayConfig holds the connection and timing details for the metrics exporter.
type PushGatewayConfig struct {
	// Endpoint is the URL of the Prometheus PushGateway (e.g., "http://localhost:9091").
	Endpoint string `json:"endpoint"`

	// Interval defines how frequently the metrics are pushed to the gateway.
	Interval utils.JSONDuration `json:"interval"`
}

// defaultConfig provides the "out-of-the-box" settings if no external configuration is found.
var defaultConfig = Config{
	BuiltInCollectors: utils.JSONStrings{
		"go_collector",                 // Go runtime stats (GC, Goroutines)
		"process_collector",            // OS process stats (CPU, Memory)
		"db_default",                   // Standard database connection pool stats
		"api_requests_total",           // HTTP/gRPC request counter
		"api_requests_latency_seconds", // Request duration histogram
		"api_request_size_bytes",       // Inbound payload size
		"api_response_size_bytes",      // Outbound payload size
	},
	PushGateway: PushGatewayConfig{
		// Default to pushing every 5 seconds.
		Interval: utils.JSONDuration{Duration: 5 * time.Second},
	},
}

// GetConfig retrieves the metrics configuration by merging defaults with
// values from the global config manager (e.g., from a YAML file or ETCD).
func GetConfig() (Config, error) {
	conf := defaultConfig
	// Uses the standardized prefix "asjard.metrics" defined in the constant package.
	if err := config.GetWithUnmarshal(constant.ConfigMetricsPrefix, &conf); err != nil {
		return conf, err
	}
	return conf.complete(), nil
}

// complete is a helper method that merges the built-in collectors with
// any additional collectors specified by the user.
func (c Config) complete() Config {
	// Merges slices and ensures no duplicates.
	c.Collectors = c.BuiltInCollectors.Merge(c.Collectors)
	return c
}
