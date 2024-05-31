package asjard

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	cfgenv "github.com/asjard/asjard/core/config/sources/env"
	cfgfile "github.com/asjard/asjard/core/config/sources/file"
	"github.com/asjard/asjard/core/handler"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/core/server"
)

// Asjard .
type Asjard struct {
	// 注册的服务列表
	servers  []server.Server
	handlers map[string][]interface{}
	hm       sync.RWMutex
	startErr chan error
}

// New 入口
func New() *Asjard {
	return &Asjard{
		handlers: make(map[string][]interface{}),
		startErr: make(chan error),
	}
}

// AddHandler 添加handler用以处理请求
func (asd *Asjard) AddHandler(protocol string, handler interface{}) error {
	asd.hm.Lock()
	if _, ok := asd.handlers[protocol]; ok {
		asd.handlers[protocol] = append(asd.handlers[protocol], handler)
	} else {
		asd.handlers[protocol] = []interface{}{handler}
	}
	asd.hm.Unlock()
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

	// 其他配置加载
	if err := config.Load(-1); err != nil {
		return err
	}

	// 日志初始化
	if err := logger.Init(); err != nil {
		return err
	}

	// 配置日志初始化后的一些运行期间变量初始化
	if err := runtime.Init(); err != nil {
		return err
	}

	// handler初始化
	if err := handler.Init(); err != nil {
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
	logger.Debug("System Starting...")
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
	// 优雅退出
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP, syscall.SIGABRT)
	select {
	case s := <-quit:
		logger.Infof("system get os signal %s start exiting...", s.String())
	case err := <-asd.startErr:
		logger.Errorf("start error: %s", err)
	}
	// 系统停止
	asd.stop()
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
		logger.Debugf("Start start server '%s'", sv.Protocol())
		// 添加handler
		asd.hm.RLock()
		for _, handler := range asd.handlers[sv.Protocol()] {
			if err := sv.AddHandler(handler); err != nil {
				return fmt.Errorf("server '%s' add handler fail[%s]", sv.Protocol(), err.Error())
			}
		}
		asd.hm.RUnlock()
		// 补全服务实例详情
		if err := svc.AddEndpoint(&server.Endpoint{
			Protocol:  sv.Protocol(),
			Addresses: sv.ListenAddresses(),
		}); err != nil {
			return fmt.Errorf("server '%s' add endpoint fail[%s]", sv.Protocol(), err.Error())
		}
		// 启动服务
		if err := sv.Start(asd.startErr); err != nil {
			return fmt.Errorf("start server '%s' fail[%s]", sv.Protocol(), err.Error())
		}
		logger.Debugf("server '%s' started", sv.Protocol())
	}
	return nil
}

// stop 系统停止
func (asd *Asjard) stop() {
	logger.Info("start remove instance from registry")
	// 从注册中心删除服务
	if err := registry.Unregiste(); err != nil {
		logger.Errorf("unregiste from registry fail[%s]", err.Error())
	}
	logger.Info("start stop server")
	for _, server := range asd.servers {
		if server.Enabled() {
			logger.Infof("server '%s' stopped", server.Protocol())
			server.Stop()
		}
	}
	// 配置中心断开连接
	config.DisConnect()
	bootstrap.Stop()
	logger.Info("system exit done")
}
