/*
Package cli provides a command-line argument parser that implements the
core/config/Source interface.

By implementing the Source interface, this package allows the framework
to treat command-line flags as a structured configuration provider.
This is essential for:
 1. Dynamic Overrides: Overriding specific settings (like port or log level)
    without modifying configuration files.
 2. Environment Injection: Passing secrets or runtime parameters in
    containerized environments (Docker/K8s).

The implementation typically maps flags to configuration keys, supporting
the framework's unified configuration access pattern.
*/
package cli

const (
	// Name is the unique identifier for this configuration source.
	Name = "cli"

	// Priority defines the precedence of this source.
	// A value of 1 usually indicates a very high priority, ensuring that
	// CLI arguments override values found in YAML files or environment variables.
	Priority = 1
)
