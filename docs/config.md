> 文件配置，支持yaml,json文件格式配置，其中变量名称使用驼峰式

## 配置所在目录

如果配置了环境变量`ASJARD_CONF_DIR`则读取该目录及子目录下的所有文件

否则读取环境变量`ASJARD_HOME_DIR`的值并拼接`conf`目录,读取该目录下及子目录下的所有文件

如果以上两个环境变量都没有设置,则读取`可执行程序`平级目录下的`conf`目录下及子目录下的所有文件

## 已支持文件格式

- [x] yaml,yml
- [ ] json
- [ ] ini
- [ ] prop,properties

## 配置说明

### 缓存配置

```yaml

```

### 加解密配置

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

### 客户端配置

### 配置中心配置

### 数据库配置

### 拦截器配置

### 日志配置

### 注册发现配置

### 协议配置

### 服务配置

```yml

```
