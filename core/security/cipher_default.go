package security

// DefaultCipher 默认加解密组件
type DefaultCipher struct {
	name string
}

// NewDefaultCipher 初始化默认加解密组件
func NewDefaultCipher(name string) (Cipher, error) {
	return &DefaultCipher{
		name: name,
	}, nil
}

// Encrypt 加密实现
func (c *DefaultCipher) Encrypt(data string, options *Options) (string, error) {
	return data, nil
}

// Decrypt 解密实现
func (c *DefaultCipher) Decrypt(data string, options *Options) (string, error) {
	return data, nil
}
