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
	// I18nInterceptorName is the unique identifier for this interceptor.
	I18nInterceptorName = "i18n"
	// localsConfDirName is the default directory name for translation files.
	localsConfDirName = "locals"
	// I18N_DIR_ENV_NAME is the environment variable to override the config path.
	I18N_DIR_ENV_NAME = "ASJARD_I18N_DIR"
	// HeaderLang is the HTTP header key used to detect client language (e.g., "lang: en-US").
	HeaderLang = "lang"
)

// I18n handles the translation of error responses.
type I18n struct {
	enabled bool
	watcher *fsnotify.Watcher
	// locals stores mapping: [language_code][error_code] -> TranslationConfig
	locals map[string]map[uint32]*I18nConfig
	lm     sync.RWMutex
}

// I18nConfig represents the structure of a translation entry.
type I18nConfig struct {
	Prompt string `json:"prompt"` // Human-readable message for end-users.
	Doc    string `json:"doc"`    // Link or detailed technical documentation.
}

func init() {
	// Register i18n support specifically for the REST protocol.
	server.AddInterceptor(I18nInterceptorName, NewI18nInterceptor, rest.Protocol)
}

// NewI18nInterceptor initializes the translation engine and starts the file watcher.
func NewI18nInterceptor() (server.ServerInterceptor, error) {
	logger.Debug("new i18 interceptor")
	// Only initialize if enabled in the configuration.
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

	// Start background watcher for hot-reloading JSON files.
	go i18n.watch()
	i18n.watcher.Add(confDir)

	// Initial load of all translation files in the directory.
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

// Interceptor returns the middleware that translates error responses.
func (m *I18n) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		// 1. Execute the handler logic first.
		resp, err = handler(ctx, req)
		if err == nil {
			return resp, err // No error, no translation needed.
		}

		if !m.enabled {
			return resp, err
		}

		// 2. Identify the protocol. Currently only REST supports this i18n flow.
		rtx, ok := ctx.(*rest.Context)
		if !ok {
			return resp, err
		}

		// 3. Extract the 'lang' header from the request.
		lang := string(rtx.Request.Header.Peek(HeaderLang))
		if lang == "" {
			return resp, err
		}

		// 4. Convert the error to an Asjard Status to get the internal Error Code.
		stts := status.FromError(err)
		conf, ok := m.getConf(lang, stts.ErrCode)
		if !ok {
			return resp, err
		}

		// 5. Wrap the localized prompt and doc into a protobuf Any detail.
		detail, _ := anypb.New(&statuspb.Status{
			Doc:    conf.Doc,
			Prompt: conf.Prompt,
		})

		// 6. Return a new gRPC-compatible status with the localized details attached.
		return resp, grpcstatus.ErrorProto(&spb.Status{
			Code:    int32(stts.Code),
			Message: stts.Message,
			Details: []*anypb.Any{detail},
		})
	}
}

// setConf thread-safely updates the translation map for a language.
func (m *I18n) setConf(lang string, conf map[uint32]*I18nConfig) {
	m.lm.Lock()
	m.locals[lang] = conf
	m.lm.Unlock()
}

// removeConf removes a language's translations (used when a file is deleted).
func (m *I18n) removeConf(lang string) {
	m.lm.Lock()
	delete(m.locals, lang)
	m.lm.Unlock()
}

// getConf retrieves a specific translation based on language and error code.
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

// load reads a JSON file and parses it into the translation map.
// The filename (without extension) is treated as the language identifier (e.g., en-US.json).
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

// remove handles the cleanup when a translation file is deleted from the disk.
func (m *I18n) remove(path string) {
	fileName := filepath.Base(path)
	ext := filepath.Ext(fileName)
	lang := strings.TrimSuffix(fileName, ext)
	m.removeConf(lang)
}

// watch listens for file system events (Create, Write, Delete) to enable hot-reloading.
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

// getI18nDir determines where translation files are stored.
func getI18nDir() string {
	dir := os.Getenv(I18N_DIR_ENV_NAME)
	if dir != "" {
		return dir
	}
	return filepath.Join(utils.GetHomeDir(), localsConfDirName)
}
