/*
Package runtime 系统运行时一些参数，系统启动时初始化，后续可直接从这里获取
*/
package runtime

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/google/uuid"
)

const (
	website = "https://github.com/%s/%s"
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
	Favicon string `json:"favicon"`
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
		Favicon:     "favicon.ico",
		Instance: Instance{
			SystemCode: 100,
			Name:       constant.Framework,
			Version:    "1.0.0",
			MetaData:   make(map[string]string),
		},
	}
	// 框架退出信号
	Exit    = make(chan struct{})
	appOnce sync.Once
)

// GetAPP 获取项目详情
// 需要配置加载完后才能加载
func GetAPP() APP {
	if !config.IsLoaded() {
		panic("config not loaded")
	}
	appOnce.Do(func() {
		if err := config.GetWithUnmarshal(constant.ConfigServicePrefix, &app); err != nil {
			logger.Error("get instance conf fail", "err", err)
		}
		if app.Instance.SystemCode < 100 || app.Instance.SystemCode > 999 {
			app.Instance.SystemCode = 100
		}
		app.Instance.ID = uuid.NewString()
		app.Website = fmt.Sprintf(website, app.App, app.Instance.Name)
		constant.APP.Store(app.App)
		constant.Region.Store(app.Region)
		constant.AZ.Store(app.AZ)
		constant.Env.Store(app.Environment)
		constant.ServiceName.Store(app.Instance.Name)
		logger.Debug("get app", "app", app)
	})
	return app
}

var bufPool = sync.Pool{
	New: func() any {
		b := bytes.NewBuffer(make([]byte, 0, 128))
		b.Reset()
		return b
	},
}

// ResourceKey 资源key
// 比如缓存中的key
// {app}:{resource}:{env}:{service}:{version}:{region}:{az}:{key}
// resource: 资源类型, 比如caches, lock
// key: 资源key
func (app APP) ResourceKey(resource, key string, opts ...Option) string {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	if resource == "" {
		resource = "resource"
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	write := func(s string) {
		if buf.Len() == 0 && options.startWithDelimiter {
			buf.WriteString(options.delimiter)
		}
		buf.WriteString(s)
		buf.WriteString(options.delimiter)
	}

	if !options.withoutApp {
		write(app.App)
	}
	write(resource)
	if !options.withoutEnv {
		write(app.Environment)
	}
	if !options.withoutService {
		if options.withServiceId {
			write(app.Instance.ID)
		} else {
			write(app.Instance.Name)
		}
	}
	if !options.withoutVersion {
		write(app.Instance.Version)
	}
	if !options.withoutRegion {
		write(app.Region)
	}
	if !options.withoutAz {
		write(app.AZ)
	}
	if key != "" {
		write(key)
	}

	if !options.endWithDelimiter {
		buf.Truncate(buf.Len() - len(options.delimiter))
	}

	s := buf.String()
	bufPool.Put(buf)
	return s
}
