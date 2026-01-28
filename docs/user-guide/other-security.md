> 加解密

## 配置

```yaml
asjard:
  ## 加解密相关配置
  cipher:
    ## 默认加解密组件名称
    default: default
    ## 如果需要加解密配置文件则需要将此配置加载到环境变量中, 例如
    ## asjard_cipher_aesCBCPkcs5padding_base64Key=
    ## asjard_cipher_aesCBCPkcs5padding_base64Iv=
    ## key为自定义加解密组件的名称,
    ## value为加解密组件需要的配置
    aesCBCPkcs5padding:
      ## 密钥， 长度必须为16,24,32
      base64Key: ""
      ## 偏移量, 长度必须为16
      ## 如果为空则获取base64Key的前16个字符
      base64Iv: ""
```

## 自定义加解密

> 实现如下方法

```go
// Cipher 加解密需要实现的接口
type Cipher interface {
	// 加密方法
	Encrypt(data string, opts *Options) (string, error)
	// 解密方法
	Decrypt(data string, opts *Options) (string, error)
}

import "github.com/asjard/asjard/core/sercurity"

// 注册加解密
func init() {
	security.AddCipher("XXXCipher", NewXXXCipher)
}

// 实现以上interface接口
func NewXXXCipher(name string) (Cipher, error) {
	// TODO
	return &XXXCipher{}, nil
}
```

## 使用

```go
import "github.com/asjard/asjard/core/sercurity"

// 自定义加解密组件加密
security.Encrypt("plain text", security.WithCipherName("XXXCipher"))
// 自定义加解密组件解密
security.Decrypt("secret text", security.WithCipherName("XXXCipher")
```

### 配置使用加解密

```go
// 获取配置并使用XXXCipher解密
config.GetString("xxx.encryptKey", config.WithCipher("XXXCipher"))
// 添加配置并使用XXXCipher加密
config.Set("xxx.encryptKey", "plain text", config.WithCipher("XXXCipher")
```
