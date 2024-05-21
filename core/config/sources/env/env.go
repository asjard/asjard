package env

import (
	"os"
	"strings"

	"github.com/asjard/asjard/core/config"
	"github.com/spf13/cast"
)

const (
	// Name 名称
	Name = "env"
	// Priority 优先级
	Priority = 0
)

// Env 环境变量配置
type Env struct {
	cb func(*config.Event)
}

func init() {
	config.AddSource(Name, Priority, New)
}

// New .
func New() (config.Sourcer, error) {
	return &Env{}, nil
}

// GetAll get all environment
func (s *Env) GetAll() map[string]*config.Value {
	configmap := make(map[string]*config.Value)
	for _, value := range os.Environ() {
		keyValue := strings.Split(value, "=")
		if len(keyValue) >= 2 {
			envKey := strings.Trim(strings.ReplaceAll(keyValue[0], "_", "."), ".")
			if envKey != "" {
				envValue := strings.Join(keyValue[1:], "")
				configmap[envKey] = &config.Value{
					Sourcer: s,
					Value:   envValue,
				}
			}
		}
	}
	return configmap
}

// Set .
func (s *Env) Set(key string, value any) error {
	envKey := strings.ReplaceAll(key, ".", "_")
	if err := os.Setenv(envKey, cast.ToString(value)); err != nil {
		return err
	}
	s.cb(&config.Event{
		Type: config.EventTypeCreate,
		Key:  envKey,
		Value: &config.Value{
			Sourcer: s,
			Value:   value,
		},
	})
	return nil
}

// Watch .
func (s *Env) Watch(cb func(*config.Event)) error {
	s.cb = cb
	return nil
}

// DisConnect 停止监听
func (s *Env) DisConnect() {}

// Priority 返回优先级
func (s *Env) Priority() int {
	return Priority
}

// Name 配置源名称
func (s *Env) Name() string {
	return Name
}
