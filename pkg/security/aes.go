/*
Package security provides cryptographic implementations for the framework.
This file specifically implements the AES-CBC-PKCS5Padding encryption scheme.
*/
package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/security"
)

const (
	// AESCipherName is the unique identifier for this encryption component.
	AESCipherName = "aesCBCPkcs5padding"
)

// AESCipher handles AES encryption and decryption operations.
type AESCipher struct {
	name  string
	block cipher.Block // The underlying AES block cipher (AES-128, 192, or 256)
	key   []byte       // The secret key used for encryption
	iv    []byte       // Initialization Vector to ensure unique ciphertexts
}

func init() {
	// Register this cipher implementation into the security manager.
	security.AddCipher(AESCipherName, NewAESCipher)
}

// NewAESCipher initializes the cipher by pulling the Key and IV from global configuration.
// Keys can be generated using: openssl rand -base64 32
func NewAESCipher(name string) (security.Cipher, error) {
	cipher, err := MustNewAESCipher(config.GetString(fmt.Sprintf(constant.ConfigCipherAESKey, name), ""),
		config.GetString(fmt.Sprintf(constant.ConfigCipherAESIV, name), ""))
	if err != nil {
		return nil, err
	}
	cipher.name = name
	return cipher, nil
}

// MustNewAESCipher creates a new AES cipher instance from base64 encoded strings.
func MustNewAESCipher(base64Key, base64Iv string) (*AESCipher, error) {
	if base64Key == "" {
		return nil, errors.New("aes base64Key not found")
	}

	// Decode the base64 key into raw bytes.
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("base64 decode '%s' value fail[%s]", base64Key, err.Error())
	}

	// Initialize the AES block cipher. The key size determines AES-128, 192, or 256.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new aes cipher fail[%s]", err.Error())
	}

	var iv []byte
	if base64Iv == "" {
		// If no IV is provided, default to using a slice of the key (not recommended for production).
		iv = key[:block.BlockSize()]
	} else {
		iv, err = base64.StdEncoding.DecodeString(base64Iv)
		if err != nil {
			return nil, fmt.Errorf("base64 decode '%s' value fail[%s]", base64Iv, err.Error())
		}
	}

	// IV length must match the AES block size (16 bytes).
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("iv length must be %d", block.BlockSize())
	}

	return &AESCipher{
		block: block,
		key:   key,
		iv:    iv,
	}, nil
}

// Encrypt encrypts a plain text string and returns a base64 encoded cipher text.
func (c *AESCipher) Encrypt(data string, options *security.Options) (string, error) {
	out, err := c.encrypt([]byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt decodes base64 data and decrypts it back into plain text.
func (c *AESCipher) Decrypt(base64Data string, options *security.Options) (string, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", err
	}
	out, err := c.decrypt(data)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Internal raw encryption logic.
func (c *AESCipher) encrypt(origData []byte) ([]byte, error) {
	// AES is a block cipher; data must be a multiple of the block size.
	origData = c.pKCS5Padding(origData, c.block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(c.block, c.iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// Internal raw decryption logic.
func (c *AESCipher) decrypt(crypted []byte) ([]byte, error) {
	if len(crypted)%c.block.BlockSize() != 0 {
		return nil, errors.New("fake encrypted data")
	}
	blockMode := cipher.NewCBCDecrypter(c.block, c.iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)

	// Remove the padding added during encryption.
	origData, err := c.pKCS5UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, nil
}

// pKCS5Padding adds bytes to the end of the plaintext to fill the final AES block.
func (c *AESCipher) pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// pKCS5UnPadding removes the PKCS5 padding from decrypted data.
func (c *AESCipher) pKCS5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length < 1 {
		return nil, errors.New("fake encrypted data")
	}
	// The value of the last byte tells us how many padding bytes were added.
	unpadding := int(origData[length-1])
	if length < unpadding {
		return nil, errors.New("fake encrypted data")
	}
	return origData[:(length - unpadding)], nil
}
