package runtime

import (
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"

	"github.com/google/uuid"
)

var (
	// ServiceID 服务ID
	ServiceID string
	// APP 项目名称
	APP string
	// Region 所属区域
	Region string
	// AZ 可用区
	AZ string
	// Environment 所属环境
	Environment string
	// Name 服务名称
	Name string
	// Version 服务版本
	Version string

	once sync.Once
)

// Init 运行期间的参数初始化
// 服务一旦启动起来后，这些参数是不会修改的
func Init() error {
	logger.Debug("Start init runtime")
	defer logger.Debug("init runtime Done")
	once.Do(func() {
		ServiceID = uuid.NewString()
		APP = config.GetString("app", constant.Framework)
		Region = config.GetString("region", "")
		AZ = config.GetString("avaliablezone", "")
		Environment = config.GetString("environment", "dev")
		Version = config.GetString("instance.version", "1.0.0")
		Name = config.GetString("instance.name", constant.Framework)
	})
	return nil
}
