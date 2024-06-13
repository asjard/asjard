package security

import (
	"testing"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/core/config/sources/env"
)

func TestAES(t *testing.T) {
	config.Load(-1)
	cipher, err := NewAESCipher(AESCipherName)
	if err != nil {
		t.Error(err.Error())

		t.FailNow()
	}
	out, err := cipher.Encrypt(`
---
testAES: xs.yml`, nil)
	if err != nil {
		t.Error(err.Error())

		t.FailNow()
	}
	t.Log(out)
}
