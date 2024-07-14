package security

import (
	"os"
	"testing"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/core/config/sources/env"
)

func TestAES(t *testing.T) {
	os.Setenv("asjard_cipher_aesCBCPkcs5padding.base64Key", "eipGxQelg9ka5oRQXy0cPQ==")
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

func TestMustNewAESCipher(t *testing.T) {
	datas := []struct {
		base64Key string
		base64Iv  string
		ok        bool
	}{
		// key不存在
		{ok: false},
		// key需要base64编码
		{base64Key: "string", ok: false},
		// 无效的key长度
		{base64Key: "XLjd8dGAt+WUFEFkxX7ga6U=", ok: false},
		// iv必须要base64编码
		{base64Key: "oroTIS4m9ln063+e1QFzlA==", base64Iv: "string", ok: false},
		// iv 长度必须为16
		{base64Key: "oroTIS4m9ln063+e1QFzlA==", base64Iv: "UH2BtzVoCXubp41yI6HC7dE=", ok: false},
		// 不带iv
		{base64Key: "oroTIS4m9ln063+e1QFzlA==", base64Iv: "", ok: true},
		// 带iv
		{base64Key: "oroTIS4m9ln063+e1QFzlA==", base64Iv: "ZFJCJrujPcPkATsRyinZWw==", ok: true},
		// key长度24
		{base64Key: "u+me+VHoiCWQGacpjlOlSgS9ROrgLwMy", base64Iv: "ZFJCJrujPcPkATsRyinZWw==", ok: true},
		{base64Key: "u+me+VHoiCWQGacpjlOlSgS9ROrgLwMy", base64Iv: "", ok: true},
		// key长度32
		{base64Key: "vc4V0yi3cpUR+VuYiqcIBw1PVBdMt2q0ZsLv8rzYXMc=", base64Iv: "ZFJCJrujPcPkATsRyinZWw==", ok: true},
		{base64Key: "vc4V0yi3cpUR+VuYiqcIBw1PVBdMt2q0ZsLv8rzYXMc=", base64Iv: "", ok: true},
	}
	for _, data := range datas {
		_, err := MustNewAESCipher(data.base64Key, data.base64Iv)
		if (err == nil) != data.ok {
			t.Errorf("key: %s, iv: %s, current: %v", data.base64Key, data.base64Iv, err)
			t.FailNow()
		}
	}
}
