package utils

import (
	"fmt"

	"github.com/asjard/asjard/core/constant"
	"github.com/magiconair/properties"
	"gopkg.in/yaml.v2"
)

// ConvertYamlToProperties yaml内容转化为props格式
/*
yaml格式转换为properties格式

yaml内容:

	a: 1
	b:
	  c: 1
	  d: [1, 2]
	  e:
	  - f: 3
		g: 4

解析后的内容应该为:

	a=1
	b.c=1
	b.d[0]=1
	b.d[1]=2
	b.e[0].f=3
	b.e[0].g=4
*/
func ConvertYamlToProperties(yamlContent []byte) (map[string]any, error) {
	ms := yaml.MapSlice{}
	if err := yaml.Unmarshal(yamlContent, &ms); err != nil {
		return nil, fmt.Errorf("yaml unmarshal fail[%s]", err.Error())
	}
	configs := make(map[string]any)
	if err := doConvertYamlToProperties("", ms, configs); err != nil {
		return nil, err
	}
	return configs, nil
}

// ConvertJsonToProperties json格式转换为properties格式
func ConvertJsonToProperties(_ []byte) (map[string]any, error) {
	return nil, nil
}

// ConvertPropsToProperties props格式转换为props
func ConvertPropsToProperties(propsContent []byte) (map[string]any, error) {
	props, err := properties.Load(propsContent, properties.UTF8)
	if err != nil {
		return nil, err
	}
	configs := make(map[string]any)
	for key, value := range props.Map() {
		configs[key] = value
	}
	return configs, nil
}

func doConvertYamlToProperties(prefix string, mapSlice yaml.MapSlice, configs map[string]any) error {
	if prefix != "" {
		prefix += constant.ConfigDelimiter
	}
	for _, item := range mapSlice {
		key, ok := item.Key.(string)
		if !ok {
			continue
		}
		switch item.Value.(type) {
		case yaml.MapSlice:
			if err := doConvertYamlToProperties(prefix+key, item.Value.(yaml.MapSlice), configs); err != nil {
				return err
			}
		case []any:
			if err := convertYamlToPropertiesWithSlice(prefix+key, item.Value.([]any), configs); err != nil {
				return err
			}
		default:
			configs[prefix+key] = item.Value
		}
	}
	return nil
}

func convertYamlToPropertiesWithSlice(prefix string, items []any, configs map[string]any) error {
	for index, value := range items {
		listKey := fmt.Sprintf("%s[%d]", prefix, index)
		switch value.(type) {
		case yaml.MapSlice:
			if err := doConvertYamlToProperties(listKey, value.(yaml.MapSlice), configs); err != nil {
				return err
			}
		case []any:
			if err := convertYamlToPropertiesWithSlice(listKey, value.([]any), configs); err != nil {
				return err
			}
		default:
			configs[listKey] = value
		}
	}
	return nil
}
