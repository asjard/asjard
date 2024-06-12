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
	AESKeyName = "cipher.aesCBCPkcs5padding.base64Key"
	AESIVName  = "cipher.aesCBCPkcs5padding.base64Iv"
)

// AESCipher aes加解密
type AESCipher struct {
	block cipher.Block
	key   []byte
	iv    []byte
}

func init() {
	security.AddCipher(AESCipherName, NewAESCipher)
}

// NewAESCipher 加解密初始化
func NewAESCipher() (security.Cipher, error) {
	// openssl rand -base64 16
	// openssl rand -base64 24
	// openssl rand -base64 32
	keyStr := config.GetString(AESKeyName, "")
	if keyStr == "" {
		return nil, fmt.Errorf("config %s not found", AESCipherName)
	}
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("base64 decode %s value fail[%s]", AESKeyName, err.Error())
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new aes cipher fail[%s]", err.Error())
	}
	ivStr := config.GetString(AESIVName, "")
	var iv []byte
	if ivStr == "" {
		iv = key[:block.BlockSize()]
	} else {
		iv, err = base64.StdEncoding.DecodeString(ivStr)
		if err != nil {
			return nil, fmt.Errorf("base64 decode %s value fail[%s]", AESIVName, err.Error())
		}
		if len(iv) != block.BlockSize() {
			return nil, fmt.Errorf("%s length must be %d", AESIVName, block.BlockSize())
		}
	}
	return &AESCipher{
		block: block,
		key:   key,
		iv:    iv,
	}, nil
}

// Encrypt 加密
func (c *AESCipher) Encrypt(data string, options *security.Options) (string, error) {
	out, err := c.encrypt([]byte(data))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Decrypt 解密
func (c *AESCipher) Decrypt(data string, options *security.Options) (string, error) {
	out, err := c.decrypt([]byte(data))
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
