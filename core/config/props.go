package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/asjard/asjard/core/constant"
	"github.com/magiconair/properties"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// PropsDecodeFunc 不同格式的数据统一转换为map[string]any类型
type PropsDecodeFunc func(content []byte, configMap map[string]any) error

var (
	yamlDecodeFunc = func(content []byte, configMap map[string]any) error {
		return yaml.Unmarshal(content, &configMap)
	}

	jsonDecodeFunc = func(content []byte, configMap map[string]any) error {
		return json.Unmarshal(content, &configMap)
	}
	propsDecodeFunc = func(content []byte, configMap map[string]any) error {
		props, err := properties.Load(content, properties.UTF8)
		if err != nil {
			return err
		}
		for key, value := range props.Map() {
			configMap[key] = value
		}
		return nil
	}
	tomlDecodeFunc = func(content []byte, configMap map[string]any) error {
		return toml.Unmarshal(content, &configMap)
	}
	propsDecodes = map[string]PropsDecodeFunc{
		"yaml":       yamlDecodeFunc,
		"yml":        yamlDecodeFunc,
		"json":       jsonDecodeFunc,
		"props":      propsDecodeFunc,
		"properties": propsDecodeFunc,
		"toml":       tomlDecodeFunc,
	}
)

// ConvertToProperties 不同格式的内容化为props格式的map
func ConvertToProperties(ext string, content []byte) (map[string]any, error) {
	decodeFunc, ok := propsDecodes[strings.ToLower(strings.Trim(ext, "."))]
	if !ok {
		return nil, fmt.Errorf("unsupport ext '%s' to convert to props", ext)
	}
	configMap := make(map[string]any)
	if err := decodeFunc(content, configMap); err != nil {
		return nil, err
	}
	propsMap := make(map[string]any)
	if err := map2Props("", configMap, propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}

// IsExtSupport 文件扩展是否支持
func IsExtSupport(ext string) bool {
	_, ok := propsDecodes[strings.ToLower(strings.Trim(ext, "."))]
	return ok
}

func map2Props(prefix string, configMap, propsMap map[string]any) error {
	if prefix != "" {
		prefix += constant.ConfigDelimiter
	}
	for key, val := range configMap {
		switch value := val.(type) {
		case map[string]any:
			if err := map2Props(prefix+key, value, propsMap); err != nil {
				return err
			}
		case []any:
			if err := slice2Props(prefix+key, value, propsMap); err != nil {
				return err
			}
		default:
			propsMap[prefix+key] = value
		}
	}
	return nil
}

func slice2Props(prefix string, items []any, propsMap map[string]any) error {
	for index, val := range items {
		listKey := fmt.Sprintf("%s[%d]", prefix, index)
		switch value := val.(type) {
		case map[string]any:
			if err := map2Props(listKey, value, propsMap); err != nil {
				return err
			}
		case []any:
			if err := slice2Props(listKey, value, propsMap); err != nil {
				return err
			}
		default:
			propsMap[listKey] = value
		}
	}
	return nil
}
