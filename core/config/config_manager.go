/*
Package config 维护各个配置源上报的配置，提供读取配置的方法和通知配置发生变更的事件
*/
package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/security"
	"github.com/asjard/asjard/utils"
	ccast "github.com/asjard/asjard/utils/cast"

	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
)

const ()

var (
	configParamCompile = regexp.MustCompile("\\${(.*?)}")
)

// Sourcer 配置源需要实现的方法
type Sourcer interface {
	// 获取所有配置,首次初始化完毕后会去配置源获取一次所有配置,
	// 维护在config_manager的本地内存中,
	// 返回的配置应该为properties格式的，并区分大小写。
	// 返回值可以通过ConvertToProperties方法获取
	GetAll() map[string]*Value
	// 添加配置到配置源中,
	// 慎用,存在安全隐患和配置源实现复杂问题
	// 理论只应该在mem配置源中使用,非必要不要使用
	Set(key string, value any) error
	// 监听配置变化,当配置源中的配置发生变化时,
	// 通过此回调方法通知config_manager进行配置变更
	Watch(func(event *Event)) error
	// 和配置中心断开连接
	Disconnect()
	// 配置中心的优先级
	Priority() int
	// 配置源名称
	Name() string
}

// NewSourceFunc 初始化配置源的方法，
// 无需携带任何参数，配置源的加载顺序是从低到高加载的，
// 高优先级的配置源在初始化的时候可以读取到低优先级的配置。
//
//	@return Sourcer 配置源
//	@return error
type NewSourceFunc func() (Sourcer, error)

// Source 配置源结构，添加的配置源保存成如下结构，
// 后续加载配置源时从此结构中读取配置源信息
type Source struct {
	name          string
	priority      int
	newSourceFunc NewSourceFunc
	// 是否已加载
	loaded bool
}

// ConfigManager 全局配置管理，
// 维护全局配置和配置源配置，
// 后续读取全局配置或者配置源配置从此结构中读取
type ConfigManager struct {
	// 配置源列表
	sourcers map[string]Sourcer
	sm       sync.RWMutex
	// 配置列表，从不同数据源的配置将key保存在此处
	// 获取配置也是从此配置中获取
	globalCfgs *configs
	// 配置源的配置维护在此处，这样就不需要每个配置源维护自己配置了
	// 使用此配置是因为当一个配置删除后，需要找到一个高优先级的配置来更新configs字段
	// 如果不维护在此处，需要新增一个GetByKey方法去配置源获取
	sourceCfgs *sourcesConfigs
	// 监听配置变化
	listener *Listener
}

var (
	sources       []*Source
	configmanager *ConfigManager
)

// 初始化configmanager全局变量
func init() {
	configmanager = &ConfigManager{
		sourcers:   make(map[string]Sourcer),
		globalCfgs: newConfigs(),
		sourceCfgs: newSourcesConfigs(),
		listener:   newListener(),
	}
}

// Load 根据配置优先级加载配置
//
//	@param priority 优先级
//	@return error
func Load(priority int) error {
	return configmanager.load(priority)
}

// AddSource 添加配置源
//
//	@param name 配置源名称
//	@param priority 配置源优先级
//	@param newSourceFunc 初始化配置源的方法
//	@return error
func AddSource(name string, priority int, newSourceFunc NewSourceFunc) error {
	for _, source := range sources {
		if name == source.name {
			return fmt.Errorf("source '%s' already exist", name)
		}
		if source.priority == priority {
			return fmt.Errorf("source '%s' priority %d is same with '%s'", name, priority, source.name)
		}
	}
	sources = append(sources, &Source{
		name:          name,
		priority:      priority,
		newSourceFunc: newSourceFunc,
		loaded:        false,
	})
	// 排序
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].priority < sources[j].priority
	})
	return nil
}

// Disconnect 和配置中心断开连接
func Disconnect() {
	configmanager.disconnect()
}

// load 加载所选优先级及以下所有配置
// 优先级<0代表加载所有配置
//
//	@receiver m
//	@param priority 优先级
//	@return error
func (m *ConfigManager) load(priority int) error {
	for _, source := range sources {
		if source.loaded {
			continue
		}

		if priority >= 0 && source.priority > priority {
			break
		}
		logger.Debug("load source", "source", source.name)
		newSourcer, err := source.newSourceFunc()
		if err != nil {
			return err
		}
		m.addSourcer(newSourcer)
		// 监听配置变化
		if err := newSourcer.Watch(m.watch); err != nil {
			return err
		}
		for key, value := range newSourcer.GetAll() {
			m.watch(&Event{
				Type:  EventTypeCreate,
				Key:   key,
				Value: value,
			})
		}
		source.loaded = true
	}
	return nil
}

// 监听配置变化
func (m *ConfigManager) watch(event *Event) {
	// logger.Debug("config changed", "event", event)
	switch event.Type {
	case EventTypeCreate, EventTypeUpdate:
		m.update(event)
	case EventTypeDelete:
		m.delete(event)
	}
}

// 查询key是否已存在
// 如果不存在则添加
// 存在, 相同的源更新，不同的源判断是否大于当前优先级
// 如果优先级高于当前值则更新
func (m *ConfigManager) update(event *Event) {
	// 配置是否存在
	value, ok := m.getConfig(event.Key)
	if !ok || value.Sourcer.Name() == event.Value.Sourcer.Name() || event.Value.Sourcer.Priority() > value.Sourcer.Priority() {
		m.setConfig(event.Value.Sourcer.Name(), event.Key, event.Value)
	}
}

// 删除, 如果没有配置则返回
// 存在则判断当前配置的优先级是否大于删除的优先级
// 如果小于则删除
// 优先级排序查找第一个
func (m *ConfigManager) delete(event *Event) {
	m.sourceCfgs.del(event.Value.Sourcer.Name(), event.Key, event.Value.Ref, event.Value.Priority)
	if event.Value.Ref != "" {
		m.deleteByRef(event)
		return
	}
	if event.Key != "" {
		m.deleteByKey(event)
	}
}

// 根据key删除
func (m *ConfigManager) deleteByKey(event *Event) {
	value, ok := m.getConfig(event.Key)
	if !ok || value.Sourcer.Priority() > event.Value.Sourcer.Priority() {
		return
	}
	m.deleteAndFindNext(event.Key)
}

// 根据引用删除
func (m *ConfigManager) deleteByRef(event *Event) {
	logger.Debug("delete by ref",
		"ref", event.Value.Ref)
	for key, value := range m.globalCfgs.getAll() {
		if value.Sourcer.Name() == event.Value.Sourcer.Name() &&
			value.Ref == event.Value.Ref {
			m.deleteAndFindNext(key)
		}
	}
}

// 删除key并且找到一个最高优先级的
// 先找到一个高优先级的，找到更新，没找到删除
func (m *ConfigManager) deleteAndFindNext(key string) {
	for i := len(sources) - 1; i >= 0; i-- {
		value, ok := m.sourceCfgs.get(sources[i].name, key)
		if !ok {
			logger.Debug("find key from source not found",
				"key", key,
				"source", sources[i].name)
			continue
		}
		logger.Debug("delete key and find next from sourcer",
			"key", key,
			"source", sources[i].name)
		m.setConfig(sources[i].name, key, value)
		return
	}
	logger.Debug("delete key", "key", key)
	m.delConfig(key)
}

func (m *ConfigManager) addSourcer(sourcer Sourcer) {
	m.sm.Lock()
	m.sourcers[sourcer.Name()] = sourcer
	m.sm.Unlock()
}

func (m *ConfigManager) getSourcer(name string) (Sourcer, bool) {
	m.sm.RLock()
	source, ok := m.sourcers[name]
	m.sm.RUnlock()
	return source, ok
}

func (m *ConfigManager) delSourcer(name string) {
	m.sm.Lock()
	delete(m.sourcers, name)
	m.sm.Unlock()
}

func (m *ConfigManager) getConfig(key string) (*Value, bool) {
	return m.globalCfgs.get(key)
}

// 倒序第一个有值的值
// 相当于依次向后覆盖，找到最终值
func (m *ConfigManager) getConfigByChain(keys []string) (value *Value, ok bool) {
	for i := len(keys) - 1; i >= 0; i-- {
		value, ok = m.getConfig(keys[i])
		if ok {
			return
		}
	}
	return
}

func (m *ConfigManager) getConfigs() map[string]*Value {
	return m.globalCfgs.getAll()
}

func (m *ConfigManager) setConfig(sourceName, key string, value *Value) {
	// 更新配置源配置
	if m.sourceCfgs.set(sourceName, key, value) {
		// 更新全局配置
		m.globalCfgs.set(key, value)
		// 异步返回事件
		go m.listener.notify(&Event{Type: EventTypeUpdate, Key: key, Value: value})
	}
}

func (m *ConfigManager) delConfig(key string) {
	m.globalCfgs.del(key)
	// 异步返回事件
	go m.listener.notify(&Event{Type: EventTypeDelete, Key: key})
}

// 获取配置
func (m *ConfigManager) getValue(key string, opts *Options) any {
	if opts != nil && opts.watch != nil {
		m.listener.watch(key, opts.watch)
	}
	value, ok := m.getConfigByChain(append([]string{key}, opts.keys...))
	if !ok {
		return nil
	}
	return m.getValueWithOptions(value.Value, opts)
}

// 添加配置监听
func (m *ConfigManager) addListener(key string, opts *Options) {
	if opts != nil && opts.watch != nil {
		m.listener.watch(key, opts.watch)
	}
}

// 移除监听器
func (m *ConfigManager) removeListener(key string) {
	m.listener.remove(key)
}

// 根据前缀获取配置
func (m *ConfigManager) getValueByPrefix(prefix string, opts *Options) map[string]any {
	if opts != nil && opts.watch != nil {
		opts.watch.pattern = prefix + ".*"
		m.listener.watch(prefix, opts.watch)
	}
	out := make(map[string]any)
	configs := m.getConfigs()
	for _, p := range append([]string{prefix}, opts.keys...) {
		for key, value := range configs {
			if strings.HasPrefix(key, p) {
				out[strings.TrimPrefix(key, p+constant.ConfigDelimiter)] = m.getValueWithOptions(value.Value, opts)
			}
		}
	}
	return out
}

// 解密数据
func (m *ConfigManager) getValueWithOptions(value any, opts *Options) any {
	if opts.cipher {
		decryptedValue, err := security.Decrypt(cast.ToString(value), security.WithCipherName(opts.cipherName))
		if err != nil {
			logger.Error("decrypt fail",
				"cipher", opts.cipherName,
				"err", err.Error())
		} else {
			value = decryptedValue
		}
	}
	// 支持参数渲染
	valueStr, ok := value.(string)
	if ok && configParamCompile.MatchString(valueStr) {
		for _, matchKey := range configParamCompile.FindAllString(valueStr, -1) {
			k1 := strings.Split(matchKey, "${")
			if len(k1) == 2 {
				k2 := strings.Split(k1[1], "}")
				if len(k2) == 2 {
					valueStr = strings.Replace(valueStr, matchKey, GetString(strings.TrimSpace(k2[0]), ""), -1)
				}
			}
		}
		value = valueStr
	}
	return value
}

// 添加配置到配置源中
// 如果没有指定配置源，判断是否设置了默认配置源，
// 如果设置了默认配置源则设置到默认配置源，
// 否则设置到所有配置源
func (m *ConfigManager) setValue(key string, value any, ops *Options) error {
	setValue := value
	if ops.cipher {
		encyptedValue, err := security.Encrypt(cast.ToString(value), security.WithCipherName(ops.cipherName))
		if err != nil {
			return err
		}
		setValue = encyptedValue
	}
	if len(ops.sourceNames) == 0 {
		return m.setValueToSource(key, GetString("asjard.config.setDefaultSource", "mem"), setValue)
	}
	for _, sourceName := range ops.sourceNames {
		if err := m.setValueToSource(key, sourceName, setValue); err != nil {
			return err
		}
	}
	return nil
}

// 如果sourceName为空则表示设置所有数据源
func (m *ConfigManager) setValueToSource(key, sourceName string, value any) error {
	if sourceName == "" {
		m.sm.RLock()
		for _, sourcer := range m.sourcers {
			logger.Debug("set key to source",
				"key", key,
				"source", sourcer.Name(),
				"value", value)
			if err := sourcer.Set(key, value); err != nil {
				return err
			}
		}
		m.sm.RUnlock()
	} else {
		m.sm.RLock()
		sourcer, ok := m.getSourcer(sourceName)
		m.sm.RUnlock()
		if !ok {
			return fmt.Errorf("source '%s' not found", sourceName)
		}
		return sourcer.Set(key, value)
	}
	return nil
}

func (m *ConfigManager) disconnect() {
	m.sm.RLock()
	defer m.sm.RUnlock()
	for name, sourcer := range m.sourcers {
		logger.Info("stop config source",
			"source", name)
		sourcer.Disconnect()
	}
}

// Set 添加配置，添加在本地内存或者远程的配置中心
func Set(key string, value any, opts ...Option) error {
	return configmanager.setValue(key, value, GetOptions(opts...))
}

// Get 获取配置
// 可以从指定源获取配置
// 自动加密或解密
//
//	@param key 需要获取的配置的key
//	@param opts 获取配置的可选参数
//	@return any
func Get(key string, options *Options) any {
	return configmanager.getValue(key, options)
}

// GetWithPrefix 根据前缀获取配置
//
//	@param prefixKey 前缀key
//	@param opts 获取配置的可选参数
//	@return map 返回的值的key为properties格式的
func GetWithPrefix(prefixKey string, opts ...Option) map[string]any {
	return configmanager.getValueByPrefix(prefixKey, GetOptions(opts...))
}

// GetString 获取配置并转化为字符串类型
//
//	@param key 配置的key
//	@param defaultValue 默认值
//	@param opts 获取配置的可选参数
//	@return string
func GetString(key string, defaultValue string, opts ...Option) string {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToStringE(v)
	if err != nil {
		return defaultValue
	}
	if options.toLower {
		return strings.ToLower(value)
	}
	if options.toUpper {
		return strings.ToUpper(value)
	}
	return value
}

// GetStrings 获取配置并返回字符串列表
//
//	@param key
//	@param defaultValue
//	@param opts
//	@return []string
func GetStrings(key string, defaultValue []string, opts ...Option) []string {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	value, err := ccast.ToStringSliceE(v, options.delimiter)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetByte 获取配置并返回[]byte类型
func GetByte(key string, defaultValue []byte, opts ...Option) []byte {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToStringE(v)
	if err != nil {
		return defaultValue
	}
	return utils.String2Byte(value)
}

// GetBool 获取配置并转化为bool类型
// 除了true, false 布尔类型外
// 如果为字符串，转换为小写字符后,
// 非 "0", "f", "false", "n", "no", "off" 均为true
// 如果为整形，不等于零均为true
//
//	@param key
//	@param defaultValue
//	@param opts
//	@return bool
func GetBool(key string, defaultValue bool, opts ...Option) bool {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	// 如果判断错误字符串转bool类型会出现错误
	// 比如不在true和false列表中的会返回false,并返回错误
	value, _ := ccast.ToBoolE(v)
	return value
}

// GetBools TODO 获取配置并转化为[]bool类型
//
//	@param key
//	@param defaultValue
//	@param opts 可选参数WithDelimiter, 默认分隔符空格
//	@return []bool
func GetBools(key string, defaultValue []bool, opts ...Option) []bool {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	valueStrs, err := ccast.ToStringSliceE(v, options.delimiter)
	if err != nil {
		return defaultValue
	}
	var value []bool
	for _, v := range valueStrs {
		vi, _ := ccast.ToBoolE(v)
		value = append(value, vi)
	}
	return value
}

// GetInt 获取配置并转化为int类型
//
//	@param key
//	@param defaultValue
//	@param opts
//	@return int
func GetInt(key string, defaultValue int, opts ...Option) int {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToIntE(v)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetInts TODO 获取配置并转化为[]int类型
//
//	@param key
//	@param defaultValue
//	@param opts 可选参数 WithDelimiter
//	@return []int
func GetInts(key string, defaultValue []int, opts ...Option) []int {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	valueStrs, err := ccast.ToStringSliceE(v, options.delimiter)
	if err != nil {
		return defaultValue
	}
	var value []int
	for _, v := range valueStrs {
		vi, err := cast.ToIntE(v)
		if err != nil {
			return defaultValue
		}
		value = append(value, vi)
	}
	return value
}

// GetInt64 获取配置并转化为int64类型
//
//	@param key
//	@param defaultValue
//	@param opts
//	@return int64
func GetInt64(key string, defaultValue int64, opts ...Option) int64 {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToInt64E(v)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetInt64s TODO 获取配置并转化为[]int64类型
func GetInt64s(key string, defaultValue []int64, opts ...Option) []int64 {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	valueStrs, err := ccast.ToStringSliceE(v, options.delimiter)
	if err != nil {
		return defaultValue
	}
	var value []int64
	for _, v := range valueStrs {
		vi, err := cast.ToInt64E(v)
		if err != nil {
			return defaultValue
		}
		value = append(value, vi)
	}
	return value
}

// GetInt32 获取配置并转化为int32类型
func GetInt32(key string, defaultValue int32, opts ...Option) int32 {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToInt32E(v)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetInt32s TODO 获取配置并转化为[]int32类型
func GetInt32s(key string, defaultValue []int32, opts ...Option) []int32 {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	valueStrs, err := ccast.ToStringSliceE(v, options.delimiter)
	if err != nil {
		return defaultValue
	}
	var value []int32
	for _, v := range valueStrs {
		vi, err := cast.ToInt32E(v)
		if err != nil {
			return defaultValue
		}
		value = append(value, int32(vi))
	}
	return value
}

// GetFloat64 获取配置并转化为float64
func GetFloat64(key string, defaultValue float64, opts ...Option) float64 {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToFloat64E(v)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetFloat64s 获取配置并转化为[]float64类型
func GetFloat64s(key string, defaultValue []float64, opts ...Option) []float64 {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	valueStrs, err := ccast.ToStringSliceE(v, options.delimiter)
	if err != nil {
		return defaultValue
	}
	var value []float64
	for _, v := range valueStrs {
		vi, err := cast.ToFloat64E(v)
		if err != nil {
			return defaultValue
		}
		value = append(value, vi)
	}
	return value
}

// GetFloat32 获取配置并转化为float32
func GetFloat32(key string, defaultValue float32, opts ...Option) float32 {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToFloat32E(v)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetFloat32s 获取配置并转化为[]float32类型
func GetFloat32s(key string, defaultValue []float32, opts ...Option) []float32 {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	valueStrs, err := ccast.ToStringSliceE(v, options.delimiter)
	if err != nil {
		return defaultValue
	}
	var value []float32
	for _, v := range valueStrs {
		vi, err := cast.ToFloat32E(v)
		if err != nil {
			return defaultValue
		}
		value = append(value, vi)
	}
	return value
}

// GetDuration 获取配置并转化为time.Duration类型
func GetDuration(key string, defaultValue time.Duration, opts ...Option) time.Duration {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToDurationE(v)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetTime 获取并转化为time.Time类型
func GetTime(key string, defaultValue time.Time, opts ...Option) time.Time {
	options := GetOptions(opts...)
	v := Get(key, options)
	if v == nil {
		return defaultValue
	}
	value, err := cast.ToTimeInDefaultLocationE(v, options.location)
	if err != nil {
		return defaultValue
	}
	return value
}

// Exist 配置是否存在
func Exist(key string) bool {
	options := GetOptions()
	return Get(key, options) != nil
}

// GetAndUnmarshal 获取结果并序列化
func GetAndUnmarshal(key string, outPtr any, opts ...Option) error {
	options := GetOptions(opts...)
	if options.unmarshaler != nil {
		return options.unmarshaler.Unmarshal([]byte(GetString(key, "", opts...)), outPtr)
	}
	return GetAndJsonUnmarshal(key, outPtr, opts...)
}

// GetAndJsonUnmarshal 获取配置并JSON反序列化
func GetAndJsonUnmarshal(key string, outPtr any, opts ...Option) error {
	return json.Unmarshal([]byte(GetString(key, "", opts...)), outPtr)
}

// GetAndYamlUnmarshal 获取配置并YAML反序列化
func GetAndYamlUnmarshal(key string, outPtr any, opts ...Option) error {
	return yaml.Unmarshal([]byte(GetString(key, "", opts...)), outPtr)
}

/*
GetWithUnmarshal 根据前缀序列化
@param prefix
@param outPtr
@param opts
@return error
*/
func GetWithUnmarshal(prefix string, outPtr any, opts ...Option) error {
	options := GetOptions(opts...)
	if options.unmarshaler != nil {
		outByte, err := json.Marshal(getConfigMap(GetWithPrefix(prefix, opts...)))
		if err != nil {
			return err
		}
		return options.unmarshaler.Unmarshal(outByte, outPtr)
	}
	return GetWithJsonUnmarshal(prefix, outPtr, opts...)
}

// GetWithJsonUnmarshal 获取key并序列化
func GetWithJsonUnmarshal(prefix string, outPtr any, opts ...Option) error {
	configMap := getConfigMap(GetWithPrefix(prefix, opts...))
	outBytes, err := json.Marshal(&configMap)
	if err != nil {
		return err
	}
	return json.Unmarshal(outBytes, outPtr)
}

// GetWithYamlUnmarshal 获取key并序列化
func GetWithYamlUnmarshal(prefix string, outPtr any, opts ...Option) error {
	configMap := getConfigMap(GetWithPrefix(prefix, opts...))
	outBytes, err := yaml.Marshal(&configMap)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(outBytes, outPtr)
}

// GetKeyValueToMap 解析配置到map中
// key=a[0]  value=1
// {"a":[1]}
func GetKeyValueToMap(key string, value any) map[string]any {
	return getConfigMap(map[string]any{
		key: value,
	})
}

// AddListener 添加配置监听
func AddListener(key string, callback func(*Event)) {
	options := GetOptions(WithWatch(callback))
	configmanager.addListener(key, options)
}

// AddPatternListener 添加匹配监听
func AddPatternListener(pattern string, callback func(*Event)) {
	options := GetOptions(WithMatchWatch(pattern, callback))
	configmanager.addListener("", options)
}

// AddPrefixListener 添加前缀监听
func AddPrefixListener(prefix string, callback func(*Event)) {
	options := GetOptions(WithPrefixWatch(prefix, callback))
	configmanager.addListener("", options)
}

// RemoveListener 移除监听器
func RemoveListener(key string) {
	configmanager.removeListener(key)
}

// 将props格式的map展开
func getConfigMap(configs map[string]any) map[string]any {
	result := make(map[string]any)
	skipKeys := make(map[string]struct{})
	for key, value := range configs {
		keyList := strings.Split(key, constant.ConfigDelimiter)
		if _, ok := skipKeys[key]; ok {
			continue
		}
		mergeConfigMap(getConfigValue(0, keyList, value, configs, skipKeys), result)
	}
	return result
}

func mergeConfigMap(from, to map[string]any) {
	for key, value := range from {
		if _, ok := to[key]; ok {
			switch v := value.(type) {
			case map[string]any:
				mergeConfigMap(v, to[key].(map[string]any))
			default:
				logger.Warn("merge fail, invalid value type want map[string]any",
					"from", from,
					"to", to,
					"value", value)
			}
		} else {
			to[key] = value
		}
	}
}

func getConfigValue(index int, keyList []string, value any, keyValue map[string]any, skipKeys map[string]struct{}) map[string]any {
	key := keyList[index]
	if index == len(keyList)-1 {
		if !strings.HasSuffix(key, "]") {
			return map[string]any{
				key: value,
			}
		}
		key = strings.Split(key, "[")[0]
		listKey := key
		preKey := strings.Join(keyList[:index], constant.ConfigDelimiter)
		if preKey != "" {
			listKey = preKey + constant.ConfigDelimiter + listKey
		}
		listIndex := 0
		var listValues []any
		for {
			indexKey := fmt.Sprintf("%s[%d]", listKey, listIndex)
			listValue, ok := keyValue[indexKey]
			if !ok {
				break
			}
			listValues = append(listValues, listValue)
			skipKeys[indexKey] = struct{}{}
			listIndex++
		}
		return map[string]any{
			key: listValues,
		}
	}
	// 列表
	if strings.HasSuffix(key, "]") {
		key = strings.Split(key, "[")[0]
		preKey := strings.Join(keyList[:index], constant.ConfigDelimiter)
		listKey := key
		if preKey != "" {
			listKey = preKey + constant.ConfigDelimiter + listKey
		}
		listIndex := 0
		var listValues []map[string]any
		for {
			listKeyValue := make(map[string]any)
			indexKey := fmt.Sprintf("%s[%d]", listKey, listIndex)
			exist := false
			for k, v := range keyValue {
				if strings.HasPrefix(k, indexKey) {
					listKeyValue[strings.Join(strings.Split(k, constant.ConfigDelimiter)[index+1:], constant.ConfigDelimiter)] = v
					skipKeys[k] = struct{}{}
					exist = true
				}
			}
			if !exist {
				break
			}
			listValues = append(listValues, getConfigMap(listKeyValue))
			listIndex++
		}
		return map[string]any{
			key: listValues,
		}
	}
	return map[string]any{
		key: getConfigValue(index+1, keyList, value, keyValue, skipKeys),
	}
}
