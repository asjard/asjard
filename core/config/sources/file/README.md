> 文件配置

- 文件扩展名表示被读取的格式如: `yaml`,`yml`,`json`等，支持扩展列表:
  - [x]yaml
  - [x]yml
  - [ ]json
  - [ ]props
  - [ ]properties
  - [ ]ini
- 首次加载以文件名asic码顺序从小到大排序进行文件读取
- 加密文件命名格式为: `encrypted_{encryptMethod}_{fileName}.{ext}`, 其中:
  - `encrypted`: 表示此文件被加密, 固定字符串，区分大小写
  - `encryptMethod`: 表示加密方式, 例如: `base64`, `aes`等
  - `fileName`: 文件名称
  - `ext`: 文件扩展名, 例如: `yaml`, `json`等