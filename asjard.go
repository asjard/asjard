package asjard

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	cfgenv "github.com/asjard/asjard/core/config/sources/env"
	cfgfile "github.com/asjard/asjard/core/config/sources/file"
	cfgmem "github.com/asjard/asjard/core/config/sources/mem"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/metrics"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/server/handlers"
	"github.com/asjard/asjard/utils"
)

const (
	website = "https://github.com/asjard/asjard"
	// http://patorjk.com/software/taag/#p=display&f=Small%20Slant&t=ASJARD
	banner = `
                                       App:      %s
                                       Env:      %s
    _   ___    _  _   ___ ___          Region:   %s
   /_\ / __|_ | |/_\ | _ \   \         Az:       %s
  / _ \\__ \ || / _ \|   / |) |
 /_/ \_\___/\__/_/ \_\_|_\___/ %s
                                       ID:       %s
                                       Name:     %s
                                       Version:  %s
                                       Website:  %s
                                       Servers:  %s
                                       ConfDir:  %s
 `
)

// Asjard .
type Asjard struct {
	// 注册的服务列表
	servers []server.Server
	// 每个协议的handlers
	handlers map[string][]any
	hm       sync.RWMutex
	// 服务启动需在后台启动，如果启动出错通过此channel返回错误
	startErr chan error
	// 已启动的服务
	startedServers []string
	// 退出信号
	exit chan struct{}
}

// New 框架初始化
func New() *Asjard {
	return &Asjard{
		handlers: make(map[string][]any),
		startErr: make(chan error),
		exit:     make(chan struct{}),
	}
}

// AddHandler 同AddHandler方法，可以让一个handler支持多个协议
func (asd *Asjard) AddHandler(handler any, protocols ...string) error {
	asd.hm.Lock()
	defer asd.hm.Unlock()
	for _, protocol := range protocols {
		if _, ok := asd.handlers[protocol]; ok {
			asd.handlers[protocol] = append(asd.handlers[protocol], handler)
		} else {
			asd.handlers[protocol] = []any{handler}
		}
		if bootstrapHandler, ok := handler.(bootstrap.BootstrapHandler); ok {
			bootstrap.AddBootstrap(bootstrapHandler)
		}
	}
	return nil
}

// Start 系统启动
func (asd *Asjard) Start() error {
	if err := asd.init(); err != nil {
		return err
	}
	if err := asd.startServers(); err != nil {
		return err
	}
	// 注册服务
	if err := registry.Registe(); err != nil {
		return err
	}
	logger.Info("System Started")
	if config.GetBool(constant.ConfigLoggerBannerEnabled, true) {
		asd.printBanner()
	}
	// 优雅退出
	quit := make(chan os.Signal)
	defer close(quit)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case s := <-quit:
		logger.Info("system get os signal start exiting...",
			"signal", s.String())
	case err := <-asd.startErr:
		logger.Error("start error:",
			"error", err)
	}
	close(asd.exit)
	// 系统停止
	asd.stop()
	return nil
}

// Exit 退出信号
// 如果系统退出则会触发此信号, 可以在stream请求或其他地方监听此信号，用以平滑退出服务
func (asd *Asjard) Exit() <-chan struct{} {
	return asd.exit
}

// 系统初始化
func (asd *Asjard) init() error {
	// 环境变量配置源加载
	if err := config.Load(cfgenv.Priority); err != nil {
		return err
	}

	// 安全组件初始化
	// 需要支持配置文件加密，所以在加载配置文件前需要加载加解密组件
	if err := security.Init(); err != nil {
		return err
	}

	// 文件配置源加载
	if err := config.Load(cfgfile.Priority); err != nil {
		return err
	}

	// 内存配置源加载
	// config.Set方法默认使用内存配置源,系统运行期间存在，系统退出后消失
	if err := config.Load(cfgmem.Priority); err != nil {
		return err
	}

	// 其他配置加载
	if err := config.Load(-1); err != nil {
		return err
	}

	// 监控初始化
	if err := metrics.Init(); err != nil {
		return err
	}

	// 客户端初始化
	if err := client.Init(); err != nil {
		return err
	}

	// 服务初始化
	servers, err := server.Init()
	if err != nil {
		return err
	}
	asd.servers = servers

	// 注册中心初始化
	if err := registry.Init(); err != nil {
		return err
	}

	// 系统启动
	if err := bootstrap.Start(); err != nil {
		return err
	}

	return nil
}

// 启动所有服务
func (asd *Asjard) startServers() error {
	logger.Debug("Start start servers")
	defer logger.Debug("start servers Done")
	svc := server.GetInstance()
	// 启动所有服务
	for _, sv := range asd.servers {
		if !sv.Enabled() {
			continue
		}
		logger.Debug("Start start server",
			"protocol", sv.Protocol())
		// 添加handler
		asd.hm.RLock()
		for _, handler := range asd.handlers[sv.Protocol()] {
			if err := sv.AddHandler(handler); err != nil {
				return fmt.Errorf("server '%s' add handler fail[%s]", sv.Protocol(), err.Error())
			}
		}
		asd.hm.RUnlock()
		// 添加默认handler
		for _, handler := range handlers.GetServerDefaultHandlers(sv.Protocol()) {
			if err := sv.AddHandler(handler); err != nil {
				return fmt.Errorf("server %s add default handler fail[%s]", sv.Protocol(), err.Error())
			}
		}

		// 补全服务实例详情
		if err := svc.AddEndpoint(sv.Protocol(), sv.ListenAddresses()); err != nil {
			return fmt.Errorf("server '%s' add endpoint fail[%s]", sv.Protocol(), err.Error())
		}
		// 启动服务
		if err := sv.Start(asd.startErr); err != nil {
			return fmt.Errorf("start server '%s' fail[%s]", sv.Protocol(), err.Error())
		}
		asd.startedServers = append(asd.startedServers, sv.Protocol())
		logger.Debug("server started",
			"protocol", sv.Protocol())
	}
	return nil
}

// stop 系统停止
func (asd *Asjard) stop() {
	logger.Info("start remove instance from registry")
	// 从注册中心删除服务
	if err := registry.Unregiste(); err != nil {
		logger.Error("unregiste from registry fail",
			"error", err.Error())
	}
	for _, server := range asd.servers {
		if server.Enabled() {
			logger.Info("start stop server", "protocol", server.Protocol())
			server.Stop()
			logger.Info("server stopped", "protocol", server.Protocol())
		}
	}
	// 配置中心断开连接
	config.Disconnect()
	bootstrap.Stop()
	logger.Info("system exited")
}

func (asd *Asjard) printBanner() {
	app := runtime.GetAPP()
	fmt.Printf(banner,
		app.App,
		app.Environment,
		app.Region,
		app.AZ,
		constant.FrameworkVersion,
		app.Instance.ID,
		app.Instance.Name,
		app.Instance.Version,
		app.Website,
		strings.Join(asd.startedServers, ","),
		utils.GetConfDir())
}
