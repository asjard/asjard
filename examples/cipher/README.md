> 加解密示例

## 配置文件

- `conf_example/examples/example_cipher.yml`
- `conf_example/examples/encrypted_aesCBCPkcs5padding_example_cipher.yml`

## 运行

> 当前目录执行

```sh
## 配置文件解密密钥
export cipher_aesCBCPkcs5padding_base64Key='L5XqRRJPm8KLI+EcqUuHHg=='
export cipher_aesCBCPkcs5padding_base64Iv='xKXiw8E1TaXJmjHL6D9+TA=='

## 启动服务
ASJARD_CONF_DIR=${PWD}/../../conf_example go run main.go
```

**查看结果**

```sh
curl -s 127.0.0.1:8080/v1/examples/cipher
```

**输出结果**

```json
{
  "code": 0,
  "message": "",
  "doc": "",
  "data": {
    "aes_encrypt_value_in_plain_file": "this is a aes encrypt value",
    "base64_encrypt_value_in_plain_file": "this is a base64 encrypt value",
    "plain_value_in_aes_encrypt_file": "this is a plain value in aes encrypted file",
    "aes_encrypt_value_in_aes_encrypt_file": "this is a aes encrypt value use another key in aes encrypted file",
    "base64_encrypt_value_in_aes_encrypt_file": "this is a base64 encrypt value in aes encrypt file"
  }
}
```
