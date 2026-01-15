package security

const (
	// DefaultCipherName is the identifier for the fallback cipher component.
	DefaultCipherName = "default"
)

// DefaultCipher is a transparent implementation of the Cipher interface.
// It performs no actual transformation on the data.
type DefaultCipher struct {
	name string
}

// NewDefaultCipher is the factory function that initializes the DefaultCipher.
// It is used by the cipherManager when no other specific cipher is configured.
func NewDefaultCipher(name string) (Cipher, error) {
	return &DefaultCipher{
		name: name,
	}, nil
}

// Encrypt returns the input string without any modification.
// This allows the system to function normally even if encryption is disabled.
func (c *DefaultCipher) Encrypt(data string, options *Options) (string, error) {
	return data, nil
}

// Decrypt returns the input string without any modification.
// It acts as the inverse of the transparent Encrypt method.
func (c *DefaultCipher) Decrypt(data string, options *Options) (string, error) {
	return data, nil
}
