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
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/server/handlers"
	"github.com/asjard/asjard/utils"
)

const (
	website = "https//github.com/asjard/asjard"
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

// New 入口
func New() *Asjard {
	return &Asjard{
		handlers: make(map[string][]any),
		startErr: make(chan error),
		exit:     make(chan struct{}),
	}
}

// AddHandler 添加handler用以处理请求
func (asd *Asjard) AddHandler(protocol string, handler any) error {
	asd.hm.Lock()
	if _, ok := asd.handlers[protocol]; ok {
		asd.handlers[protocol] = append(asd.handlers[protocol], handler)
	} else {
		asd.handlers[protocol] = []any{handler}
	}
	asd.hm.Unlock()
	if bootstrapHandler, ok := handler.(bootstrap.BootstrapHandler); ok {
		bootstrap.AddBootstrap(bootstrapHandler)
	}
	return nil
}

func (asd *Asjard) AddHandlerV2(handler any, protocols ...string) error {
	for _, protocol := range protocols {
		if err := asd.AddHandler(protocol, handler); err != nil {
			return err
		}
	}
	return nil
}

// 系统初始化
func (asd *Asjard) init() error {
	// 环境变量配置加载
	if err := config.Load(cfgenv.Priority); err != nil {
		return err
	}

	// 安全组件初始化
	if err := security.Init(); err != nil {
		return err
	}

	// 文件配置加载
	if err := config.Load(cfgfile.Priority); err != nil {
		return err
	}

	// 内存配置源加载
	if err := config.Load(cfgmem.Priority); err != nil {
		return err
	}

	// 其他配置加载
	if err := config.Load(-1); err != nil {
		return err
	}

	// 一些运行期间变量初始化
	if err := runtime.Init(); err != nil {
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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case s := <-quit:
		logger.Info("system get os signal start exiting...",
			"signal", s.String())
	case err := <-asd.startErr:
		logger.Error("start error:",
			"error", err)
	}
	asd.exit <- struct{}{}
	close(asd.exit)
	close(quit)
	// 系统停止
	asd.stop()
	return nil
}

// Exit 退出信号
func (asd *Asjard) Exit() chan struct{} {
	return asd.exit
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
	fmt.Printf(banner,
		runtime.APP,
		runtime.Environment,
		runtime.Region,
		runtime.AZ,
		constant.FrameworkVersion,
		runtime.InstanceID,
		runtime.Name,
		runtime.Version,
		config.GetString(constant.ConfigWebsite, website),
		strings.Join(asd.startedServers, ","),
		utils.GetConfDir())
}
