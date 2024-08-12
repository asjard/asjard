package interceptors

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/protobuf/statuspb"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard/utils"
	"github.com/fsnotify/fsnotify"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	I18nInterceptorName = "i18n"
	// 配置存放路径
	localsConfDirName = "locals"
	// I18N_DIR_ENV_NAME i18配置存放路径环境变量名称
	I18N_DIR_ENV_NAME = "ASJARD_I18N_DIR"
	HeaderLang        = "lang"
)

// I18n i18n拦截器
type I18n struct {
	enabled bool
	watcher *fsnotify.Watcher
	locals  map[string]map[uint32]*I18nConfig
	lm      sync.RWMutex
}

// I18nConfig i18n配置
type I18nConfig struct {
	Prompt string `json:"prompt"`
	Doc    string `json:"doc"`
}

func init() {
	// 支持rest协议
	server.AddInterceptor(I18nInterceptorName, NewI18nInterceptor, rest.Protocol)
}

// I18n拦截器初始化
func NewI18nInterceptor() (server.ServerInterceptor, error) {
	logger.Debug("new i18 interceptor")
	if !config.GetBool("asjard.servers.rest.i18n.enabled", false) {
		return &I18n{}, nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	i18n := &I18n{
		enabled: true,
		watcher: watcher,
		locals:  make(map[string]map[uint32]*I18nConfig),
	}
	confDir := getI18nDir()
	if !utils.IsPathExists(confDir) {
		return nil, fmt.Errorf("path %s not exist", confDir)
	}
	go i18n.watch()
	i18n.watcher.Add(confDir)
	if err := filepath.Walk(confDir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			i18n.load(path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return i18n, nil
}

func (*I18n) Name() string {
	return I18nInterceptorName
}

func (m *I18n) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err == nil {
			return resp, err
		}
		if !m.enabled {
			return resp, err
		}
		rtx, ok := ctx.(*rest.Context)
		if !ok {
			return resp, err
		}
		lang := rtx.GetHeaderParam(HeaderLang)
		if len(lang) == 0 {
			return resp, err
		}
		stts := status.FromError(err)
		conf, ok := m.getConf(lang[0], stts.Code)
		if !ok {
			return resp, err
		}
		detail, _ := anypb.New(&statuspb.Status{
			Doc:    conf.Doc,
			Prompt: conf.Prompt,
		})
		return resp, grpcstatus.ErrorProto(&spb.Status{
			Code:    int32(stts.Code),
			Message: stts.Message,
			Details: []*anypb.Any{detail},
		})
	}
}

func (m *I18n) setConf(lang string, conf map[uint32]*I18nConfig) {
	m.lm.Lock()
	m.locals[lang] = conf
	m.lm.Unlock()
}

func (m *I18n) removeConf(lang string) {
	m.lm.Lock()
	delete(m.locals, lang)
	m.lm.Unlock()
}

func (m *I18n) getConf(lang string, code uint32) (*I18nConfig, bool) {
	m.lm.RLock()
	defer m.lm.RUnlock()
	langConf, ok := m.locals[lang]
	if !ok {
		return nil, false
	}
	conf, ok := langConf[code]
	return conf, ok
}

func (m *I18n) load(path string) {
	logger.Debug("load i18n", "path", path)
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName)
	lang := strings.TrimSuffix(fileName, ext)
	conf := make(map[uint32]*I18nConfig)
	content, err := os.ReadFile(path)
	if err != nil {
		logger.Error("read file fail", "file", path, "err", err)
		return
	}
	switch ext {
	case ".json":
		if err := json.Unmarshal(content, &conf); err != nil {
			logger.Error("unmarshal file fail", "file", path, "err", err)
			return
		}
		m.setConf(lang, conf)
	}
}

func (m *I18n) remove(path string) {
	logger.Debug("remove i18n", "path", path)
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName)
	lang := strings.TrimSuffix(fileName, ext)
	m.removeConf(lang)
}

func (m *I18n) watch() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			switch event.Op {
			case fsnotify.Create, fsnotify.Write:
				m.load(event.Name)
			case fsnotify.Remove, fsnotify.Rename:
				m.remove(event.Name)
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("watch err", "err", err)
		}
	}
}

func getI18nDir() string {
	dir := os.Getenv(I18N_DIR_ENV_NAME)
	if dir != "" {
		return dir
	}
	return filepath.Join(utils.GetHomeDir(), localsConfDirName)
}
