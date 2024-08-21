package security

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	base64Cipher, err := NewBase64Cipher(Base64CipherName)
	assert.Nil(t, err)
	plainText := "test_plain_text"
	t.Run("Encrypt", func(t *testing.T) {
		encodeText, err := base64Cipher.Encrypt(plainText, &Options{})
		assert.Nil(t, err)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(plainText)), encodeText)
	})
	t.Run("Decrypt", func(t *testing.T) {
		t.Run("InvalidBase64", func(t *testing.T) {
			_, err := base64Cipher.Decrypt("invalid base64 data", &Options{})
			assert.NotNil(t, err)
		})
		t.Run("Success", func(t *testing.T) {
			encodeText := base64.StdEncoding.EncodeToString([]byte(plainText))
			decodeText, err := base64Cipher.Decrypt(encodeText, &Options{})
			assert.Nil(t, err)
			assert.Equal(t, plainText, decodeText)
		})
	})
}
