package config

import "fmt"

// Value represents a single configuration entry enriched with metadata.
// It tracks not just the data itself, but also its origin and priority context.
type Value struct {
	// Sourcer identifies the specific configuration provider (e.g., ETCD, File, Apollo)
	// that supplied this value. This allows the manager to track source-specific state.
	Sourcer Sourcer

	// Value is the actual configuration data. It can be any primitive type or a slice/map,
	// though it is typically normalized to a string or basic type during ingestion.
	Value any

	// Ref (Reference) is a grouping identifier used for bulk lifecycle management.
	// For example, if a source is a directory of files, Ref might be the filename.
	// This allows the manager to delete all keys associated with a file even if
	// the specific keys are no longer known at the time of deletion.
	Ref string

	// Priority defines the precedence of this specific value.
	// Even within a single source, different values might have different priorities
	// (e.g., a CLI flag vs. a default value within the same internal provider).
	Priority int
}

// String provides a formatted string representation of the Value object.
// This is primarily used for logging and debugging configuration overrides.
func (v Value) String() string {
	if v.Sourcer != nil {
		return fmt.Sprintf("sourcer: '%s', value: '%+v', ref: '%s'",
			v.Sourcer.Name(), v.Value, v.Ref)
	}
	return fmt.Sprintf("sourcer: 'nil', value: '%+v', ref: '%s'",
		v.Value, v.Ref)
}
