package config

import (
	"time"

	"github.com/asjard/asjard/core/constant"
)

// Unmarshaler defines the interface for custom deserialization.
// This allows converting configuration data into complex Go structs using JSON, YAML, or Proto.
type Unmarshaler interface {
	Unmarshal(data []byte, v any) error
}

// Options defines the behavior parameters for reading or writing configurations.
type Options struct {
	// watch specifies if and how the application should be notified of changes to this config.
	watch *watchOptions
	// sourceNames defines specific configuration providers (e.g., "apollo", "etcd") to target.
	// If empty, operations usually default to all registered sources or the default source.
	sourceNames []string
	// cipherName identifies the security component used for manual encryption/decryption.
	cipherName string
	// cipher indicates whether the manual encryption/decryption pipeline is active.
	cipher bool
	// disableAutoDecryptValue prevents the system from automatically decrypting values
	// that start with the 'encrypted_' prefix.
	disableAutoDecryptValue bool
	// location is used when casting string values to time.Time objects.
	location *time.Location
	// delimiter is used when splitting strings into slices (e.g., "a,b,c" -> []string).
	delimiter string
	// unmarshaler provides a custom strategy for struct mapping.
	unmarshaler Unmarshaler
	// ignoreCase specifies if key lookups should treat 'Key' and 'key' as identical.
	ignoreCase bool
	// toUpper forces the retrieved string value to uppercase.
	toUpper bool
	// toLower forces the retrieved string value to lowercase.
	toLower bool
	// keys provides an ordered list of fallback keys to check if the primary key is missing.
	keys []string
}

// ListenFunc is a filter function used to determine if a callback should be triggered based on event details.
type ListenFunc func(*Event) bool

// watchOptions contains the configuration for the Observer/Listener logic.
type watchOptions struct {
	// pattern is a regex string for matching keys.
	pattern string
	// callback is the user-defined function to execute when a change occurs.
	callback CallbackFunc
	// f is an optional filter function to refine notification triggers.
	f ListenFunc
}

// clone creates a deep copy of watchOptions to prevent side effects during dynamic prefix resolution.
func (w *watchOptions) clone() *watchOptions {
	return &watchOptions{
		pattern:  w.pattern,
		callback: w.callback,
		f:        w.f,
	}
}

// Option is the functional argument type used by Get and Set methods.
type Option func(*Options)

// WithWatch attaches a simple direct-key listener.
func WithWatch(callback CallbackFunc) func(opts *Options) {
	return func(opts *Options) {
		opts.watch = &watchOptions{
			callback: callback,
		}
	}
}

// WithMatchWatch attaches a regex-based listener to the configuration request.
func WithMatchWatch(pattern string, callback CallbackFunc) func(opts *Options) {
	return func(opts *Options) {
		if pattern == "" {
			return
		}
		opts.watch = &watchOptions{
			pattern:  pattern,
			callback: callback,
		}
	}
}

// WithPrefixWatch attaches a listener that triggers for any key under a specific namespace.
func WithPrefixWatch(prefix string, callback CallbackFunc) Option {
	return func(opts *Options) {
		opts.watch = &watchOptions{
			pattern:  prefix + ".*",
			callback: callback,
		}
	}
}

// WithSource specifies a particular configuration source to interact with.
func WithSource(sourceName string) func(opts *Options) {
	return func(opts *Options) {
		// Ensure unique source names.
		for _, name := range opts.sourceNames {
			if name == sourceName {
				return
			}
		}
		opts.sourceNames = append(opts.sourceNames, sourceName)
	}
}

// WithDisableAutoDecryptValue turns off the 'encrypted_' prefix detection logic.
func WithDisableAutoDecryptValue() func(opts *Options) {
	return func(opts *Options) {
		opts.disableAutoDecryptValue = true
	}
}

// WithCipher enables security processing with a specific named cipher.
func WithCipher(cipherName string) func(opts *Options) {
	return func(opts *Options) {
		opts.cipher = true
		opts.cipherName = cipherName
	}
}

// WithLocation sets the timezone for time-based configuration values.
func WithLocation(location *time.Location) func(opts *Options) {
	return func(opts *Options) {
		opts.location = location
	}
}

// WithUnmarshaler sets the custom deserializer for complex objects.
func WithUnmarshaler(unmarshaler Unmarshaler) func(opts *Options) {
	return func(opts *Options) {
		opts.unmarshaler = unmarshaler
	}
}

// WithDelimiter sets a custom character for splitting slice configurations.
func WithDelimiter(delimiter string) func(opts *Options) {
	return func(opts *Options) {
		opts.delimiter = delimiter
	}
}

// WithIgnoreCase enables case-insensitive key searching.
func WithIgnoreCase() func(opts *Options) {
	return func(opts *Options) {
		opts.ignoreCase = true
	}
}

// WithToUpper transforms retrieved strings to uppercase.
func WithToUpper() func(opts *Options) {
	return func(opts *Options) {
		opts.toUpper = true
	}
}

// WithToLower transforms retrieved strings to lowercase.
func WithToLower() func(opts *Options) {
	return func(opts *Options) {
		opts.toLower = true
	}
}

// WithChain defines a priority-ordered sequence of keys to find a value.
func WithChain(keys []string) func(opts *Options) {
	return func(opts *Options) {
		opts.keys = keys
	}
}

// GetOptions processes a slice of Option functions and returns the final populated Options struct.
// It initializes defaults like the system-wide delimiter.
func GetOptions(opts ...Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	if options.delimiter == "" {
		options.delimiter = constant.DefaultDelimiter
	}
	return options
}
