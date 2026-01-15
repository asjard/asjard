package security

// Options encapsulates the configuration required for an encryption or decryption operation.
type Options struct {
	// cipherName specifies which registered implementation to use (e.g., "aes", "base64", "rsa").
	cipherName string
	// params stores arbitrary key-value pairs for algorithm-specific requirements.
	// For example, an AES cipher might look here for a 'key' or 'nonce'.
	params map[string]any
}

// Option is a function type used to modify the Options struct.
type Option func(opts *Options)

// WithCipherName returns an Option that sets the name of the cipher to be used.
// If the provided name is empty, the current value (usually the default) is preserved.
func WithCipherName(name string) func(opts *Options) {
	return func(opts *Options) {
		if name != "" {
			opts.cipherName = name
		}
	}
}

// WithParams returns an Option that attaches a map of custom arguments to the operation.
// This is useful for passing dynamic configuration that isn't part of the standard interface.
func WithParams(params map[string]any) func(opts *Options) {
	return func(opts *Options) {
		opts.params = params
	}
}

// Name returns the identifier of the selected cipher.
func (opts *Options) Name() string {
	return opts.cipherName
}

// Params returns the custom parameter map.
func (opts *Options) Params() map[string]any {
	return opts.params
}
