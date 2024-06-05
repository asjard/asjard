package security

import (
	"fmt"
	"sync"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
)

// Cipher 加解密需要实现的接口
type Cipher interface {
	// 加密方法
	Encrypt(data string, opts *Options) (string, error)
	// 解密方法
	Decrypt(data string, opts *Options) (string, error)
}

// NewCipherFunc 加解密组件初始化方法
type NewCipherFunc func() (Cipher, error)

type cipher struct {
	name    string
	newFunc NewCipherFunc
}

// cipherManager 加解密组件管理
type cipherManager struct {
	sync.RWMutex
	ciphers map[string]Cipher

	// 默认加解密组件名称
	defaultCipherName string
}

var (
	cpm        *cipherManager
	newCiphers []*cipher
)

func init() {
	cpm = &cipherManager{
		ciphers:           make(map[string]Cipher),
		defaultCipherName: constant.DefaultCipherName,
	}
	AddCipher(constant.DefaultCipherName, NewDefaultCipher)
}

// Init 加解密组件初始化
func Init() error {
	logger.Debug("Start init cipher")
	defer logger.Debug("init cipher done")
	for _, cph := range newCiphers {
		newCipher, err := cph.newFunc()
		if err != nil {
			return err
		}
		logger.Debugf("cipher '%s' inited", cph.name)
		cpm.add(cph.name, newCipher)
	}
	return nil
}

// AddCipher 添加加解密组件
func AddCipher(name string, newFunc NewCipherFunc) error {
	for _, cph := range newCiphers {
		if cph.name == name {
			return fmt.Errorf("cipher '%s' already exist", name)
		}
	}
	newCiphers = append(newCiphers, &cipher{
		name:    name,
		newFunc: newFunc,
	})
	logger.Debugf("cipher '%s' added", name)
	return nil
}

// GetCipher 获取加解密组件
func GetCipher(name string) (Cipher, error) {
	return cpm.get(name)
}

// Encrypt 加密内容
func Encrypt(data string, opts ...Option) (string, error) {
	return cpm.encrypt(data, opts...)
}

// Decrypt 解密内容
func Decrypt(data string, opts ...Option) (string, error) {
	return cpm.decrypt(data, opts...)
}

func (c *cipherManager) add(name string, cph Cipher) {
	c.Lock()
	c.ciphers[name] = cph
	c.Unlock()
}

func (c *cipherManager) get(name string) (Cipher, error) {
	c.RLock()
	cph, ok := c.ciphers[name]
	c.RUnlock()
	if !ok {
		return nil, fmt.Errorf("cipher '%s' not found", name)
	}
	return cph, nil
}

func (c *cipherManager) encrypt(data string, opts ...Option) (string, error) {
	options := c.options(opts...)
	cipher, err := c.get(options.Name())
	if err != nil {
		return "", err
	}
	return cipher.Encrypt(data, options)
}

func (c *cipherManager) decrypt(data string, opts ...Option) (string, error) {
	options := c.options(opts...)
	cipher, err := c.get(options.Name())
	if err != nil {
		return "", err
	}
	return cipher.Decrypt(data, options)
}

func (c *cipherManager) options(optFuncs ...Option) *Options {
	options := &Options{
		cipherName: c.defaultCipherName,
	}
	for _, optFunc := range optFuncs {
		optFunc(options)
	}
	return options
}
