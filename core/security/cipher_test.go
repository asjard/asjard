package security

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCipher struct{}

func (testCipher) Encrypt(data string, opts *Options) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

func (testCipher) Decrypt(data string, opts *Options) (string, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func newTestCipher(_ string) (Cipher, error) {
	return &testCipher{}, nil
}

func TestCipher(t *testing.T) {
	AddCipher("testCipher", newTestCipher)
	testData := "test_data"
	testCipherName := "testCipher"
	t.Run("GetCipher", func(t *testing.T) {
		tc, err := GetCipher(testCipherName)
		assert.Nil(t, err)
		assert.NotNil(t, tc)
	})
	t.Run("Encrypt", func(t *testing.T) {
		d, err := Encrypt(testData, WithCipherName(testCipherName))
		assert.Nil(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(testData)), d)
	})
	t.Run("Decrypt", func(t *testing.T) {
		d, err := Decrypt(base64.StdEncoding.EncodeToString([]byte(testData)), WithCipherName(testCipherName))
		assert.Nil(t, err)
		assert.Equal(t, testData, d)
	})
}
