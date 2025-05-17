package asjard

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/config/sources/file"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/metrics"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/server/handlers"
	"github.com/asjard/asjard/core/trace"
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
                                       Servers:  %s
                                       ConfDir:  %s
 `
)

// Asjard 维护框架所需启动的服务，以及每个服务的handler，用以在start阶段使用
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
	// 是否已初始化了
	inited atomic.Bool
}

// New 框架初始化
func New() *Asjard {
	return &Asjard{
		handlers: make(map[string][]any),
		startErr: make(chan error),
	}
}

// AddHandler 给协议添加handler，一个handler可以处理不同的协议
// 具体handler需要实现什么方法需要每个协议自行定义
// 这里只是维护不同协议的handler列表，待到start时使用
func (asd *Asjard) AddHandler(handler any, protocols ...string) error {
	asd.hm.Lock()
	defer asd.hm.Unlock()
	for _, protocol := range protocols {
		if _, ok := asd.handlers[protocol]; ok {
			asd.handlers[protocol] = append(asd.handlers[protocol], handler)
		} else {
			asd.handlers[protocol] = []any{handler}
		}
		if bootstrapHandler, ok := handler.(bootstrap.Initiator); ok {
			bootstrap.AddBootstrap(bootstrapHandler)
		}
	}
	return nil
}

// AddHandlers 功能同AddHandler方法, 添加同一个协议的多个handler
func (asd *Asjard) AddHandlers(protocol string, handlers ...any) error {
	for _, handler := range handlers {
		if err := asd.AddHandler(handler, protocol); err != nil {
			return err
		}
	}
	return nil
}

// Start 系统启动, 先根据配置初始化各个组件
func (asd *Asjard) Start() error {
	logger.Info("System Starting...")
	if err := asd.Init(); err != nil {
		return err
	}

	if err := asd.startServers(); err != nil {
		return err
	}

	// 注册服务
	if err := registry.Registe(); err != nil {
		return err
	}
	// 服务发现
	if err := registry.Discover(); err != nil {
		return err
	}
	logger.Info("System Started")
	if config.GetBool(constant.ConfigLoggerBannerEnabled, true) {
		asd.printBanner()
	}
	// 优雅退出
	quit := make(chan os.Signal, 1)
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
	close(runtime.Exit)
	// 系统停止
	asd.stop()
	return nil
}

// Exit 退出信号
// 如果系统退出则会触发此信号, 可以在stream请求或其他地方监听此信号，用以平滑退出服务
func (asd *Asjard) Exit() <-chan struct{} {
	return runtime.Exit
}

// Init 系统初始化
func (asd *Asjard) Init() error {
	if asd.inited.Load() {
		return nil
	}
	defer asd.inited.Store(true)
	// 文件配置源加载
	if err := config.Load(file.Priority); err != nil {
		return err
	}

	// 其他组件初始化之前的组件初始化
	if err := bootstrap.Init(); err != nil {
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

	// 链路追踪初始化
	if err := trace.Init(); err != nil {
		return err
	}

	// 客户端初始化
	if err := client.Init(); err != nil {
		return err
	}

	// 注册中心初始化
	if err := registry.Init(); err != nil {
		return err
	}

	// 系统启动
	if err := bootstrap.Bootstrap(); err != nil {
		return err
	}

	// 服务初始化
	servers, err := server.Init()
	if err != nil {
		return err
	}
	asd.servers = servers
	return nil
}

// 启动所有服务
func (asd *Asjard) startServers() error {
	logger.Debug("Start start servers")
	defer logger.Debug("start servers Done")
	svc := server.GetService()
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
		listenAddresses := sv.ListenAddresses()
		if err := svc.AddEndpoint(sv.Protocol(), listenAddresses); err != nil {
			return fmt.Errorf("server '%s' add endpoint fail[%s]", sv.Protocol(), err.Error())
		}
		// 启动服务
		if err := sv.Start(asd.startErr); err != nil {
			return fmt.Errorf("start server '%s' fail[%s]", sv.Protocol(), err.Error())
		}
		protocolPrefix := sv.Protocol() + "://"
		asd.startedServers = append(asd.startedServers,
			strings.TrimSuffix(strings.Join([]string{
				protocolPrefix + listenAddresses.Listen,
				protocolPrefix + listenAddresses.Advertise,
			}, ","), ","+protocolPrefix))
		logger.Debug("server started",
			"protocol", sv.Protocol())
	}
	return nil
}

// stop 系统停止
func (asd *Asjard) stop() {
	logger.Debug("start remove instance from registry")
	// 从注册中心删除服务
	if err := registry.Unregiste(); err != nil {
		logger.Error("unregiste from registry fail",
			"error", err.Error())
	}
	for _, server := range asd.servers {
		if server.Enabled() {
			logger.Debug("start stop server", "protocol", server.Protocol())
			server.Stop()
			logger.Debug("server stopped", "protocol", server.Protocol())
		}
	}
	time.Sleep(time.Second)
	// 配置中心断开连接
	config.Disconnect()
	bootstrap.Shutdown()
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
		// app.Website,
		strings.Join(asd.startedServers, ";"),
		utils.GetConfDir())
}
