> AES加解密命令行工具

## 使用帮助

### 查看帮助

```sh
asjard_cipher_aes -h
```

### 加解密文本

```sh
base64Key=$(openssl rand -base64 16)
base64Iv=$(openssl rand -base64 16)
asjard_cipher_aes -t 'hello world' -k "${base64Key}" -v "${base64Iv}"
# encrypt text SUCCESS, base64 output: Z41NvZ0EdtWsQ/hW0qbAHg==

asjard_cipher_aes -d  -t 'Z41NvZ0EdtWsQ/hW0qbAHg==' -k "${base64Key}" -v "${base64Iv}"
```

### 加解密文件

```sh
base64Key=$(openssl rand -base64 16)
base64Iv=$(openssl rand -base64 16)
asjard_cipher_aes -f ./conf_example/examples/test.yaml -k "${base64Key}" -v "${base64Iv}"
```
