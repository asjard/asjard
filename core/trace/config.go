package trace

import (
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/utils"
)

type Config struct {
	Enabled bool `json:"enabled"`
	// 带协议路径的地址
	// http://127.0.0.1:4318
	// grpc://127.0.0.1:4319
	Endpoint string             `json:"endpoint"`
	Timeout  utils.JSONDuration `json:"timeout"`
	CertFile string             `json:"certFile"`
	KeyFile  string             `json:"keyFile"`
	CaFile   string             `json:"cafile"`
}

var defaultTraceConfig = Config{
	Timeout: utils.JSONDuration{Duration: time.Second},
}

func GetConfig() *Config {
	conf := defaultTraceConfig
	config.GetWithUnmarshal("asjard.trace", &conf)
	return &conf
}
