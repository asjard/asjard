/*
Package runtime 系统运行时一些参数，系统启动时初始化，后续可直接从这里获取
*/
package runtime

import (
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/google/uuid"
)

const (
	website = "https://github.com/asjard/asjard"
)

type APP struct {
	// 所属项目
	App string `json:"app"`
	// 所属环境
	Environment string `json:"environment"`
	// 所属区域
	Region string `json:"region"`
	// 可用区
	AZ string `json:"avaliablezone"`
	// 项目站点
	Website string `json:"website"`
	// 项目描述
	Desc string `json:"desc"`
	// 实例详情
	Instance Instance `json:"instance"`
}

type Instance struct {
	// 实例ID
	ID string
	// 系统码
	SystemCode uint32 `json:"systemCode"`
	// 实例名称
	Name string `json:"name"`
	// 实例版本
	Version string `json:"version"`
	// 是否可共享
	Shareable bool `json:"shareable"`
	// 服务元数据
	MetaData map[string]string `json:"metadata"`
}

var (
	app = APP{
		App:         constant.Framework,
		Environment: "dev",
		Region:      "default",
		AZ:          "default",
		Website:     website,
		Instance: Instance{
			SystemCode: 100,
			Name:       constant.Framework,
			Version:    "1.0.0",
			MetaData:   make(map[string]string),
		},
	}
	appOnce sync.Once
)

// GetAPP 获取项目详情
func GetAPP() APP {
	appOnce.Do(func() {
		if err := config.GetWithUnmarshal(constant.ConfigServicePrefix, &app); err != nil {
			logger.Error("get instance conf fail", "err", err)
		}
		if app.Instance.SystemCode < 100 || app.Instance.SystemCode > 999 {
			app.Instance.SystemCode = 100
		}
		app.Instance.ID = uuid.NewString()
		logger.Debug("get app", "app", app)
	})
	return app
}
