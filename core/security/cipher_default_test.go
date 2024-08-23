package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	defaultCipher, err := NewDefaultCipher("default")
	assert.Nil(t, err)
	plainText := "test_plain_text"
	t.Run("Encrypt", func(t *testing.T) {
		encodeText, err := defaultCipher.Encrypt(plainText, &Options{})
		assert.Nil(t, err)
		assert.Equal(t, plainText, encodeText)
	})
	t.Run("Decrypt", func(t *testing.T) {
		decodeText, err := defaultCipher.Decrypt(plainText, &Options{})
		assert.Nil(t, err)
		assert.Equal(t, plainText, decodeText)
	})
}
