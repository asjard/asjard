package security

import (
	"encoding/base64"
)

const (
	// Base64CipherName is the unique identifier for the Base64 "cipher" component.
	Base64CipherName = "base64"
)

// Base64Cipher implements the security.Cipher interface using standard Base64.
type Base64Cipher struct {
	name string
}

func init() {
	// Automatically registers the Base64 implementation into the security manager
	// when the package is imported.
	AddCipher(Base64CipherName, NewBase64Cipher)
}

// NewBase64Cipher is the factory function that initializes a new Base64Cipher instance.
func NewBase64Cipher(name string) (Cipher, error) {
	return &Base64Cipher{
		name: name,
	}, nil
}

// Encrypt converts plaintext strings into Base64 encoded strings.
// It follows the security.Cipher interface signature, though it does not use 'options'
// as Base64 is a fixed-algorithm scheme.
func (c *Base64Cipher) Encrypt(data string, options *Options) (string, error) {
	// Standard Base64 encoding (RFC 4648).
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

// Decrypt converts Base64 encoded strings back into their original plaintext.
// Returns an error if the input string is not valid Base64 data.
func (c *Base64Cipher) Decrypt(data string, options *Options) (string, error) {
	out, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
