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

// PropsDecodeFunc defines a function signature for decoding raw bytes into a raw nested map.
type PropsDecodeFunc func(content []byte, configMap map[string]any) error

var (
	// yamlDecodeFunc handles both .yaml and .yml files.
	yamlDecodeFunc = func(content []byte, configMap map[string]any) error {
		return yaml.Unmarshal(content, &configMap)
	}

	// jsonDecodeFunc handles standard JSON content.
	jsonDecodeFunc = func(content []byte, configMap map[string]any) error {
		return json.Unmarshal(content, &configMap)
	}

	// propsDecodeFunc handles Java-style .properties files.
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

	// tomlDecodeFunc handles TOML configuration files.
	tomlDecodeFunc = func(content []byte, configMap map[string]any) error {
		return toml.Unmarshal(content, &configMap)
	}

	// propsDecodes maps file extensions to their respective decoding logic.
	propsDecodes = map[string]PropsDecodeFunc{
		"yaml":       yamlDecodeFunc,
		"yml":        yamlDecodeFunc,
		"json":       jsonDecodeFunc,
		"props":      propsDecodeFunc,
		"properties": propsDecodeFunc,
		"toml":       tomlDecodeFunc,
	}
)

// ConvertToProperties converts content from a supported format (YAML, JSON, etc.)
// into a flattened properties-style map.
// Example: A nested YAML "server: {port: 80}" becomes a map entry "server.port": 80.
func ConvertToProperties(ext string, content []byte) (map[string]any, error) {
	// Identify the decoder based on file extension.
	decodeFunc, ok := propsDecodes[strings.ToLower(strings.Trim(ext, "."))]
	if !ok {
		return nil, fmt.Errorf("unsupport ext '%s' to convert to props", ext)
	}

	// Step 1: Decode raw bytes into a nested map structure.
	configMap := make(map[string]any)
	if err := decodeFunc(content, configMap); err != nil {
		return nil, err
	}

	// Step 2: Flatten the nested map into dot-notation properties.
	propsMap := make(map[string]any)
	if err := Map2Props("", configMap, propsMap); err != nil {
		return nil, err
	}
	return propsMap, nil
}

// IsExtSupport checks if the configuration system has a decoder for the given file extension.
func IsExtSupport(ext string) bool {
	_, ok := propsDecodes[strings.ToLower(strings.Trim(ext, "."))]
	return ok
}

// Map2Props recursively flattens a nested map into a single-level map with dot-delimited keys.
// Example: {"a": {"b": 1}} -> {"a.b": 1}
func Map2Props(prefix string, configMap, propsMap map[string]any) error {
	if prefix != "" {
		prefix += constant.ConfigDelimiter // Usually "."
	}
	for key, val := range configMap {
		switch value := val.(type) {
		case map[string]any:
			// Recurse into nested maps.
			if err := Map2Props(prefix+key, value, propsMap); err != nil {
				return err
			}
		case []any:
			// Handle slices/arrays by converting them to indexed properties.
			if err := slice2Props(prefix+key, value, propsMap); err != nil {
				return err
			}
		default:
			// Leaf node: store the final value with its full path prefix.
			propsMap[prefix+key] = value
		}
	}
	return nil
}

// slice2Props handles the conversion of array items into indexed dot-notation.
// Example: "users": ["alice", "bob"] -> {"users[0]": "alice", "users[1]": "bob"}
func slice2Props(prefix string, items []any, propsMap map[string]any) error {
	for index, val := range items {
		listKey := fmt.Sprintf("%s[%d]", prefix, index)
		switch value := val.(type) {
		case map[string]any:
			// Handle objects within slices.
			if err := Map2Props(listKey, value, propsMap); err != nil {
				return err
			}
		case []any:
			// Handle nested slices (multidimensional arrays).
			if err := slice2Props(listKey, value, propsMap); err != nil {
				return err
			}
		default:
			propsMap[listKey] = value
		}
	}
	return nil
}
