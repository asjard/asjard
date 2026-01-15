/*
Package env implements the environment variable configuration source,
fulfilling the core/config/Source interface.
*/
package env

import (
	"os"
	"strings"

	"github.com/asjard/asjard/core/config"
	"github.com/spf13/cast"
)

const (
	// Name is the unique identifier for the environment variable source.
	Name = "env"
	// Priority defines the precedence of this source.
	// A value of 0 indicates the absolute highest priority,
	// meaning environment variables will override CLI flags, files, and remote configs.
	Priority = 0
)

// Env handles reading and writing configuration via OS environment variables.
type Env struct {
	options *config.SourceOptions
}

func init() {
	// Automatically register this source into the framework's configuration manager.
	config.AddSource(Name, Priority, New)
}

// New initializes the environment variable configuration source.
func New(options *config.SourceOptions) (config.Sourcer, error) {
	return &Env{
		options: options,
	}, nil
}

// GetAll scans all OS environment variables and converts them into framework-compatible keys.
// It converts underscores (_) to dots (.) to match the internal configuration hierarchy.
// Example: "ASJARD_SERVER_PORT=8080" becomes the key "asjard.server.port".
func (s *Env) GetAll() map[string]*config.Value {
	configmap := make(map[string]*config.Value)
	for _, value := range os.Environ() {
		// Environment variables are stored as "KEY=VALUE"
		keyValue := strings.SplitN(value, "=", 2)
		if len(keyValue) >= 2 {
			// Normalize the key: replace underscores with dots for internal use.
			envKey := strings.Trim(strings.ReplaceAll(keyValue[0], "_", "."), ".")
			if envKey != "" {
				envValue := strings.Join(keyValue[1:], "")
				configmap[envKey] = &config.Value{
					Sourcer: s,
					Value:   envValue,
				}
			}
		}
	}
	return configmap
}

// Set updates an environment variable and triggers a configuration change event.
// It converts internal dot-notation keys back to underscore-notation.
func (s *Env) Set(key string, value any) error {
	// Map internal "app.port" back to "APP_PORT" for the OS.
	envKey := strings.ReplaceAll(key, ".", "_")
	if err := os.Setenv(envKey, cast.ToString(value)); err != nil {
		return err
	}

	// Notify the framework that a configuration value has changed.
	s.options.Callback(&config.Event{
		Type: config.EventTypeCreate,
		Key:  key,
		Value: &config.Value{
			Sourcer: s,
			Value:   value,
		},
	})
	return nil
}

// Disconnect is a no-op for environment variables as they don't require active connections.
func (s *Env) Disconnect() {}

// Priority returns the precedence level (0 = Highest).
func (s *Env) Priority() int {
	return Priority
}

// Name returns "env".
func (s *Env) Name() string {
	return Name
}
