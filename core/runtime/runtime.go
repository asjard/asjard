package runtime

import (
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"

	"github.com/google/uuid"
)

var (
	// InstanceID 实例ID
	InstanceID string
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
	once.Do(func() {
		InstanceID = uuid.NewString()
		APP = config.GetString(constant.ConfigApp, constant.Framework)
		Region = config.GetString(constant.ConfigRegion, "default")
		AZ = config.GetString(constant.ConfigAvaliablezone, "default")
		Environment = config.GetString(constant.ConfigEnvironment, "dev")
		Version = config.GetString(constant.ConfigVersion, "1.0.0")
		Name = config.GetString(constant.ConfigName, constant.Framework)
	})
	return nil
}
