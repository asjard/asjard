package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/security"
)

const (
	// AESCipherName aes加解密组件名称
	AESCipherName = "aesCBCPkcs5padding"
	// AESKeyName 密钥key名称
	AESKeyName = "cipher.%s.base64Key"
	AESIVName  = "cipher.%s.base64Iv"
)

// AESCipher aes加解密
type AESCipher struct {
	name  string
	block cipher.Block
	key   []byte
	iv    []byte
}

func init() {
	security.AddCipher(AESCipherName, NewAESCipher)
}

// NewAESCipher 加解密初始化
// openssl rand -base64 16
// openssl rand -base64 24
// openssl rand -base64 32
func NewAESCipher(name string) (security.Cipher, error) {
	cipher, err := MustNewAESCipher(config.GetString(fmt.Sprintf(AESKeyName, name), ""),
		config.GetString(fmt.Sprintf(AESIVName, name), ""))
	if err != nil {
		return nil, err
	}
	cipher.name = name
	return cipher, nil
}

// NewAESCipherWithConfig 根据配置初始化
func MustNewAESCipher(base64Key, base64Iv string) (*AESCipher, error) {
	if base64Key == "" {
		return nil, errors.New("aes base64Key not found")
	}
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("base64 decode '%s' value fail[%s]", base64Key, err.Error())
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new aes cipher fail[%s]", err.Error())
	}
	var iv []byte
	if base64Iv == "" {
		iv = key[:block.BlockSize()]
	} else {
		iv, err = base64.StdEncoding.DecodeString(base64Iv)
		if err != nil {
			return nil, fmt.Errorf("base64 decode '%s' value fail[%s]", base64Iv, err.Error())
		}
	}
	if len(iv) != block.BlockSize() {
		return nil, fmt.Errorf("iv length must be %d", block.BlockSize())
	}
	return &AESCipher{
		block: block,
		key:   key,
		iv:    iv,
	}, nil
}

// Encrypt 加密
// 明文加密返回base64编码后的数据
func (c *AESCipher) Encrypt(data string, options *security.Options) (string, error) {
	out, err := c.encrypt([]byte(data))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt 解密
// base64编码后的秘文解密返回明文
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

func (c *AESCipher) encrypt(origData []byte) ([]byte, error) {
	origData = c.pKCS5Padding(origData, c.block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(c.block, c.iv)
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func (c *AESCipher) decrypt(crypted []byte) ([]byte, error) {
	if len(crypted)%c.block.BlockSize() != 0 {
		return nil, errors.New("fake encrypted data")
	}
	blockMode := cipher.NewCBCDecrypter(c.block, c.iv)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData, err := c.pKCS5UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, nil
}

func (c *AESCipher) pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (c *AESCipher) pKCS5UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length < 1 {
		return nil, errors.New("fake encrypted data")
	}
	unpadding := int(origData[length-1])
	if length < unpadding {
		return nil, errors.New("fake encrypted data")
	}
	return origData[:(length - unpadding)], nil
}
