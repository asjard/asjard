package security

import (
	"fmt"
	"sync"
)

// Cipher is the interface that must be implemented by any encryption component.
type Cipher interface {
	// Encrypt transforms plaintext into ciphertext based on the provided options.
	Encrypt(data string, opts *Options) (string, error)
	// Decrypt transforms ciphertext back into plaintext.
	Decrypt(data string, opts *Options) (string, error)
}

// NewCipherFunc defines the factory function signature for initializing a Cipher.
type NewCipherFunc func(name string) (Cipher, error)

// cipher internal structure to hold cipher metadata.
type cipher struct {
	name    string
	newFunc NewCipherFunc
}

// cipherManager coordinates the lifecycle and access of multiple Cipher implementations.
type cipherManager struct {
	sync.RWMutex
	// ciphers stores active, initialized Cipher instances.
	ciphers map[string]Cipher

	// defaultCipherName is used when no specific cipher is requested in Options.
	defaultCipherName string
}

var (
	// cpm is the global singleton manager for ciphers.
	cpm *cipherManager
	// newCiphers is a registry of factory functions, populated via AddCipher.
	newCiphers = make(map[string]NewCipherFunc)
	ncm        sync.RWMutex
)

func init() {
	cpm = &cipherManager{
		ciphers:           make(map[string]Cipher),
		defaultCipherName: DefaultCipherName,
	}
	// Register the framework's built-in default cipher.
	AddCipher(DefaultCipherName, NewDefaultCipher)
}

// AddCipher registers a new Cipher implementation into the global registry.
// This is typically called from an 'init' function in a sub-package.
func AddCipher(name string, newFunc NewCipherFunc) {
	ncm.Lock()
	newCiphers[name] = newFunc
	ncm.Unlock()
}

// GetCipher retrieves a specific Cipher by name, initializing it if necessary.
func GetCipher(name string) (Cipher, error) {
	return cpm.get(name)
}

// Encrypt is a high-level helper to encrypt data using functional options.
func Encrypt(data string, opts ...Option) (string, error) {
	return cpm.encrypt(data, opts...)
}

// Decrypt is a high-level helper to decrypt data using functional options.
func Decrypt(data string, opts ...Option) (string, error) {
	return cpm.decrypt(data, opts...)
}

// add stores an initialized Cipher in the manager's thread-safe map.
func (c *cipherManager) add(name string, cph Cipher) {
	c.Lock()
	c.ciphers[name] = cph
	c.Unlock()
}

// get handles the lazy-loading logic: check cache first, then initialize via factory if missing.
func (c *cipherManager) get(name string) (Cipher, error) {
	c.RLock()
	cph, ok := c.ciphers[name]
	c.RUnlock()

	if !ok {
		ncm.Lock()
		newCipherFunc, ok := newCiphers[name]
		ncm.Unlock()
		if !ok {
			return nil, fmt.Errorf("cipher '%s' not found", name)
		}

		// Initialize the new cipher instance.
		newCipher, err := newCipherFunc(name)
		if err != nil {
			return nil, err
		}

		// Cache it for future use.
		c.add(name, newCipher)
		return newCipher, nil
	}
	return cph, nil
}

// encrypt selects the appropriate cipher based on options and executes the encryption.
func (c *cipherManager) encrypt(data string, opts ...Option) (string, error) {
	options := c.options(opts...)
	cipher, err := c.get(options.Name())
	if err != nil {
		return "", err
	}
	return cipher.Encrypt(data, options)
}

// decrypt selects the appropriate cipher based on options and executes the decryption.
func (c *cipherManager) decrypt(data string, opts ...Option) (string, error) {
	options := c.options(opts...)
	cipher, err := c.get(options.Name())
	if err != nil {
		return "", err
	}
	return cipher.Decrypt(data, options)
}

// options merges user-provided functional options with the manager's defaults.
func (c *cipherManager) options(optFuncs ...Option) *Options {
	options := &Options{
		cipherName: c.defaultCipherName,
	}
	for _, optFunc := range optFuncs {
		optFunc(options)
	}
	return options
}
