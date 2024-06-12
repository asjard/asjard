package security

import (
	"encoding/base64"

	"github.com/asjard/asjard/core/security"
)

const (
	// Base64CipherName base64加密组件名称
	Base64CipherName = "base64"
)

// Base64Cipher base64加解密组件
type Base64Cipher struct{}

func init() {
	// 注册加解密组件
	security.AddCipher(Base64CipherName, NewBase64Cipher)
}

// NewBase64Cipher 初始化base64加解密组件
func NewBase64Cipher() (security.Cipher, error) {
	return &Base64Cipher{}, nil
}

// Encrypt 加密实现
func (c *Base64Cipher) Encrypt(data string, options *security.Options) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

// Decrypt 解密实现
func (c *Base64Cipher) Decrypt(data string, options *security.Options) (string, error) {
	out, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
