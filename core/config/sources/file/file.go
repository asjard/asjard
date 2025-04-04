/*
Package filesource 监听，读取，解析本地文件中的配置，
当配置发生变更时通知config_manager变更配置
*/
package file

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/utils"

	"github.com/fsnotify/fsnotify"
)

const (
	// Name 配置源名称
	Name = "file"
	// Priority 配置源优先级
	Priority = 2
	// FileNameSplitSymbol 文件名称分隔符
	// 用来区分是否为加密文件,或者其他需要解析文件名的标志
	FileNameSplitSymbol = "_"
	// FileEncryptFlag 文件加密标志
	FileEncryptFlag = "encrypted"
)

// File 文件配置源
type File struct {
	// 事件回调
	options *config.SourceOptions
	// 文件列表, 初始化时会将指定目录下的文件扫描到此列表中
	// 后续增加文件时此处不会变更，只在初始化时有用
	files []string
	// 目录列表, 和files功能相似
	// 用来监听目录下文件变化
	dirs []string
	// 记录每个文件key
	// 用来处理回调事件中的事件类型
	// 如果没有此处记录，无法在回调事件中处理删除事件
	configs map[string]map[string]any
	// 配置目录
	confDir string
	// configs的锁
	cm sync.RWMutex
	// 文件监听
	watcher *fsnotify.Watcher
}

// 初始化添加文件配置源到config_manager中
func init() {
	config.AddSource(Name, Priority, New)
}

// New 初始化文件配置源,
// 初始化需要读取的文件列表,
// 监听文件的变化.
func New(options *config.SourceOptions) (config.Sourcer, error) {
	fsource := &File{
		configs: make(map[string]map[string]any),
		options: options,
	}

	fsource.confDir = utils.GetConfDir()
	if !utils.IsPathExists(fsource.confDir) {
		return fsource, nil
	}
	if err := fsource.walk(fsource.confDir); err != nil {
		return fsource, err
	}
	if err := fsource.watch(); err != nil {
		return fsource, err
	}
	return fsource, nil
}

// GetAll config_manager初始化完毕后会调用此接口读取所有配置,
// 只有初始化完毕后调用一次,
// 后续当文件配置源中的配置发生变化后需通过watch的回调方法通知config_manager变更配置.
func (s *File) GetAll() map[string]*config.Value {
	configs := make(map[string]*config.Value)
	for _, file := range s.files {
		fileConfigs, err := s.read(file)
		if err == nil {
			for key, value := range fileConfigs {
				configs[key] = value
				s.setConfig(file, key, value.Value)
			}
		} else {
			logger.Error("read file fail",
				"file", file,
				"err", err.Error())
		}
	}
	return configs
}

// Set 设置配置到文件配置源中,暂不实现
// 本地文件中的配置应该做到只读权限，不可修改
func (s *File) Set(key string, value any) error {
	return nil
}

// DisConnect 停止监听
func (s *File) Disconnect() {
	if s.watcher != nil {
		s.watcher.Close()
	}
}

// Priority 返回配置源的优先级
func (s *File) Priority() int {
	return Priority
}

// Name 配置源名称
func (s *File) Name() string {
	return Name
}

// 读取文件内容，并解析为config_manager所需要的properties格式
func (s *File) read(file string) (map[string]*config.Value, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	baseName := filepath.Base(file)
	if strings.HasPrefix(baseName, FileEncryptFlag) {
		nameList := strings.Split(baseName, FileNameSplitSymbol)
		var decryptOptions []security.Option
		if len(nameList) > 2 {
			decryptOptions = append(decryptOptions, security.WithCipherName(nameList[1]))
		}
		decryptContent, err := security.Decrypt(string(content), decryptOptions...)
		if err != nil {
			return nil, err
		}
		content = []byte(decryptContent)
	}
	contentMap, err := config.ConvertToProperties(filepath.Ext(file), content)
	if err != nil {
		return nil, err
	}
	configs := make(map[string]*config.Value)
	for key, value := range contentMap {
		configs[key] = &config.Value{
			Sourcer: s,
			Value:   value,
			Ref:     file,
		}
	}
	return configs, nil
}

func (s *File) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	s.watcher = watcher
	go s.doWatch()
	for _, dir := range s.dirs {
		if err := s.watcher.Add(dir); err != nil {
			return err
		}
	}
	return nil
}

func (s *File) doWatch() {
	for {
		select {
		case event, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			switch event.Op {
			case fsnotify.Create, fsnotify.Write:
				logger.Debug("file source watch event",
					"event", event, "op", event.Op.String())
				if utils.IsDir(event.Name) {
					if err := s.watcher.Add(event.Name); err != nil {
						logger.Error("watch dir fail", "dir", event.Name, "err", err)
					}
				} else {
					configs, err := s.read(event.Name)
					if err == nil {
						for _, event := range s.getUpdateEvents(event.Name, configs) {
							s.options.Callback(event)
						}
					} else {
						logger.Error("read file fail",
							"file", event.Name,
							"err", err.Error())
					}
				}
			case fsnotify.Remove, fsnotify.Rename:
				logger.Debug("file source watch event",
					"event", event, "op", event.Op.String())
				if utils.IsDir(event.Name) {
					if err := s.watcher.Remove(event.Name); err != nil {
						logger.Error("remove watch dir fail", "dir", event.Name, "err", err)
					}
				} else {
					s.delConfig(event.Name, "")
					s.options.Callback(&config.Event{
						Type: config.EventTypeDelete,
						Value: &config.Value{
							Sourcer:  s,
							Ref:      event.Name,
							Priority: -1,
						},
					})
				}
			}
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("watch err",
				"err", err)
		}
	}
}

// 遍历目录
/*
文件里表保持和编辑器一致
[{"dir": ["file1", "file2"]}, {"dir": ["file1"]}]
当添加一个文件，或者删除一个文件，该文件后的所有优先级都会发生变化
*/
func (s *File) walk(dir string) error {
	if err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			s.files = append(s.files, path)
			return nil
		}
		s.dirs = append(s.dirs, path)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *File) setConfig(file, key string, value any) {
	s.cm.Lock()
	if _, ok := s.configs[file]; !ok {
		s.configs[file] = make(map[string]any)
	}
	s.configs[file][key] = value
	s.cm.Unlock()
}

func (s *File) getUpdateEvents(file string, configs map[string]*config.Value) []*config.Event {
	var events []*config.Event
	s.cm.Lock()
	defer s.cm.Unlock()
	if _, ok := s.configs[file]; !ok {
		s.configs[file] = make(map[string]any)
	}

	for key := range s.configs[file] {
		if _, ok := configs[key]; !ok {
			events = append(events, &config.Event{
				Type: config.EventTypeDelete,
				Key:  key,
				Value: &config.Value{
					Sourcer: s,
				},
			})
			delete(s.configs[file], key)
		} else {

		}
	}
	for key, value := range configs {
		if oldValue, ok := s.configs[file][key]; !ok || oldValue != value.Value {
			events = append(events, &config.Event{
				Type:  config.EventTypeUpdate,
				Key:   key,
				Value: value,
			})
			s.configs[file][key] = value.Value
		}
	}
	return events
}

func (s *File) delConfig(file, key string) {
	s.cm.Lock()
	defer s.cm.Unlock()
	if key == "" {
		delete(s.configs, file)
		return
	}
	if _, ok := s.configs[file]; ok {
		delete(s.configs[file], key)
	}
}
