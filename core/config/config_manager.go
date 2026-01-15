package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/security"
	ccast "github.com/asjard/asjard/utils/cast"

	"github.com/spf13/cast"
	"gopkg.in/yaml.v2"
)

var (
	// Regex to identify ${variable} patterns for dynamic parameter injection.
	configParamCompile = regexp.MustCompile("\\${(.*?)}")
	// Atomic flag to ensure the configuration system is initialized only once.
	loadedFlag atomic.Bool
)

const (
	// Prefix identifying a value that needs decryption.
	ValueEncryptFlag = "encrypted_"
	// Delimiter used to separate the flag from the cipher component name.
	ValueEncryptCipherNameSplitSymbol = "_"
	// Delimiter used to separate the cipher metadata from the encrypted payload.
	ValueEncryptSplitSymbol = ":"
)

// Sourcer defines the interface for external configuration providers (e.g., Apollo, ETCD, Local Files).
type Sourcer interface {
	// GetAll retrieves a full snapshot of configurations in property format (case-sensitive).
	GetAll() map[string]*Value
	// Set persists a configuration to the source. Primarily intended for memory-based sources.
	Set(key string, value any) error
	// Disconnect safely closes the connection to the configuration center.
	Disconnect()
	// Priority returns the precedence level of the source; higher values override lower ones.
	Priority() int
	// Name returns the unique identifier for the configuration source.
	Name() string
}

// NewSourceFunc is a factory type for initializing a Sourcer.
// Sources are loaded in ascending order of priority, allowing high-priority
// sources to read values from low-priority ones during initialization.
type NewSourceFunc func(options *SourceOptions) (Sourcer, error)

// Source represents a registered provider waiting to be instantiated.
type Source struct {
	name          string
	priority      int
	newSourceFunc NewSourceFunc
	// Tracks if the source has been initialized.
	loaded bool
}

// ConfigManager is the central controller managing global and source-specific state.
type ConfigManager struct {
	// Active configuration source instances.
	sourcers map[string]Sourcer
	sm       sync.RWMutex
	// The final flattened configuration map used for read operations.
	globalCfgs Configer
	// Raw values from every source, used to determine the next-best value when a key is deleted.
	sourceCfgs SourcesConfiger
	// Management of configuration change subscribers.
	listener *Listener
}

// CallbackFunc is the handler signature for configuration update events.
type CallbackFunc func(event *Event)

// SourceOptions provides context for Sourcer initialization.
type SourceOptions struct {
	Callback CallbackFunc
}

// SourceOption is a functional argument for configuring SourceOptions.
type SourceOption func(options *SourceOptions)

// WithCallback attaches an event listener to the configuration source.
func WithCallback(callback CallbackFunc) func(options *SourceOptions) {
	return func(options *SourceOptions) {
		options.Callback = callback
	}
}

// NewSourceOptions initializes SourceOptions with functional arguments.
func NewSourceOptions(opts ...SourceOption) *SourceOptions {
	options := &SourceOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

var (
	// Global registry of all available configuration sources.
	sources       []*Source
	configmanager *ConfigManager
)

// Initialize the global ConfigManager instance.
func init() {
	configmanager = &ConfigManager{
		sourcers:   make(map[string]Sourcer),
		globalCfgs: &ConfigsWithSyncMap{},
		sourceCfgs: &SourcesConfigWithSyncMap{},
		listener:   newListener(),
	}
}

// Load triggers the loading of sources up to the specified priority.
// If priority < 0, all registered sources are loaded.
func Load(priority int) error {
	defer loadedFlag.CompareAndSwap(false, true)
	return configmanager.load(priority)
}

// IsLoaded returns true if the initial configuration load has completed.
func IsLoaded() bool {
	return loadedFlag.Load()
}

// AddSource registers a new provider. Prevents duplicate names or priority levels.
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
	// Sort sources to ensure strictly deterministic override behavior.
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].priority < sources[j].priority
	})
	return nil
}

// Disconnect terminates all active configuration source connections.
func Disconnect() {
	configmanager.disconnect()
}

// load executes the initialization of Sourcers based on the priority threshold.
func (m *ConfigManager) load(priority int) error {
	for _, source := range sources {
		if source.loaded {
			continue
		}

		if priority >= 0 && source.priority > priority {
			break
		}
		logger.Debug("load source", "source", source.name)
		newSourcer, err := source.newSourceFunc(NewSourceOptions(WithCallback(m.watch)))
		if err != nil {
			return err
		}
		m.addSourcer(newSourcer)
		// Seed global config with initial snapshot from the source.
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

// watch handles configuration lifecycle events from sources.
func (m *ConfigManager) watch(event *Event) {
	switch event.Type {
	case EventTypeCreate, EventTypeUpdate:
		m.update(event)
	case EventTypeDelete:
		m.delete(event)
	}
}

// update implements the priority-based "winner-take-all" logic.
// A new value replaces the current global value if it comes from a higher-priority source.
func (m *ConfigManager) update(event *Event) {
	value, ok := m.getConfig(event.Key)
	if !ok || value.Sourcer.Name() == event.Value.Sourcer.Name() || event.Value.Sourcer.Priority() > value.Sourcer.Priority() {
		m.setConfig(event.Value.Sourcer.Name(), event.Key, event.Value)
	}
}

// delete removes a value and triggers a fallback search for the next highest priority value.
func (m *ConfigManager) delete(event *Event) {
	m.sourceCfgs.Del(event.Value.Sourcer.Name(), event.Key, event.Value.Ref, event.Value.Priority)
	if event.Value.Ref != "" {
		m.deleteByRef(event)
		return
	}
	if event.Key != "" {
		m.deleteByKey(event)
	}
}

// deleteByKey handles explicit key deletions.
func (m *ConfigManager) deleteByKey(event *Event) {
	value, ok := m.getConfig(event.Key)
	if !ok || value.Sourcer.Priority() > event.Value.Sourcer.Priority() {
		return
	}
	m.deleteAndFindNext(event.Key)
}

// deleteByRef handles deletions of all keys associated with a specific reference (e.g., a file).
func (m *ConfigManager) deleteByRef(event *Event) {
	logger.Debug("delete by ref", "ref", event.Value.Ref)
	for key, value := range m.globalCfgs.GetAll() {
		if value.Sourcer.Name() == event.Value.Sourcer.Name() &&
			value.Ref == event.Value.Ref {
			m.deleteAndFindNext(key)
		}
	}
}

// deleteAndFindNext performs a reverse-priority search to find a replacement for a deleted key.
func (m *ConfigManager) deleteAndFindNext(key string) {
	for i := len(sources) - 1; i >= 0; i-- {
		value, ok := m.sourceCfgs.Get(sources[i].name, key)
		if !ok {
			continue
		}
		m.setConfig(sources[i].name, key, value)
		return
	}
	m.delConfig(key)
}

// addSourcer registers an active sourcer instance.
func (m *ConfigManager) addSourcer(sourcer Sourcer) {
	m.sm.Lock()
	m.sourcers[sourcer.Name()] = sourcer
	m.sm.Unlock()
}

// getSourcer retrieves an active sourcer by name.
func (m *ConfigManager) getSourcer(name string) (Sourcer, bool) {
	m.sm.RLock()
	source, ok := m.sourcers[name]
	m.sm.RUnlock()
	return source, ok
}

// delSourcer removes an active sourcer.
func (m *ConfigManager) delSourcer(name string) {
	m.sm.Lock()
	delete(m.sourcers, name)
	m.sm.Unlock()
}

// getConfig gets the currently active value for a key.
func (m *ConfigManager) getConfig(key string) (*Value, bool) {
	return m.globalCfgs.Get(key)
}

// getConfigByChain checks a slice of keys and returns the first one found.
func (m *ConfigManager) getConfigByChain(keys []string) (value *Value, ok bool) {
	for i := len(keys) - 1; i >= 0; i-- {
		value, ok = m.getConfig(keys[i])
		if ok {
			return
		}
	}
	return
}

// getConfigs returns all currently active configurations.
func (m *ConfigManager) getConfigs() map[string]*Value {
	return m.globalCfgs.GetAll()
}

// getConfigsWithPrefixs returns configurations matching prefixes with the prefix trimmed from the key.
func (m *ConfigManager) getConfigsWithPrefixs(prefixs ...string) map[string]*Value {
	return m.globalCfgs.GetAllWithPrefixs(prefixs...)
}

// setConfig updates the global configuration and notifies listeners asynchronously.
func (m *ConfigManager) setConfig(sourceName, key string, value *Value) {
	if m.sourceCfgs.Set(sourceName, key, value) {
		m.globalCfgs.Set(key, value)
		go m.listener.notify(&Event{Type: EventTypeUpdate, Key: key, Value: value})
	}
}

// delConfig removes a key globally and notifies listeners asynchronously.
func (m *ConfigManager) delConfig(key string) {
	m.globalCfgs.Del(key)
	go m.listener.notify(&Event{Type: EventTypeDelete, Key: key})
}

// getValue retrieves a value with support for fallback keys and dynamic type casting.
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

// addListener registers a callback for configuration changes.
func (m *ConfigManager) addListener(key string, opts *Options) {
	if opts != nil && opts.watch != nil {
		m.listener.watch(key, opts.watch)
	}
}

// removeListener unregisters all callbacks for a specific key.
func (m *ConfigManager) removeListener(key string) {
	m.listener.remove(key)
}

// getValueByPrefix supports resolving keys where the prefix itself might be a dynamic parameter.
func (m *ConfigManager) getValueByPrefix(prefix string, opts *Options) map[string]any {
	valuePrefix := ""
	if ok, value := m.valueIsParam(GetString(prefix, "")); ok {
		valuePrefix = value
	}
	if opts != nil && opts.watch != nil {
		opts.watch.pattern = prefix + ".*"
		m.listener.watch(prefix, opts.watch)
		if valuePrefix != "" {
			pwatch := opts.watch.clone()
			pwatch.pattern = valuePrefix + ".*"
			m.listener.watch(prefix, pwatch)
		}
	}
	if valuePrefix != "" {
		prefix = valuePrefix
	}
	return m.getValueWithPrefix(prefix, opts)

}

// getValueWithPrefix recursively resolves values matching a prefix, including parameter redirection.
func (m *ConfigManager) getValueWithPrefix(prefix string, opts *Options) map[string]any {
	out := make(map[string]any)
	for key, value := range m.getConfigsWithPrefixs(append([]string{prefix}, opts.keys...)...) {
		if ok, pv := m.valueIsParam(cast.ToString(value.Value)); ok && m.getValue(pv, opts) == nil {
			for k, v := range m.getValueWithPrefix(pv, opts) {
				out[key+"."+k] = v
			}
		} else {
			out[key] = m.getValueWithOptions(value.Value, opts)
		}
	}
	return out
}

// valueIsParam checks if a string is a variable placeholder like ${key}.
func (m *ConfigManager) valueIsParam(value string) (bool, string) {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		return true, strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")
	}
	return false, value
}

// getValueWithOptions handles post-retrieval processing: manual decryption, parameter rendering, and auto-decryption.
func (m *ConfigManager) getValueWithOptions(value any, opts *Options) any {
	// 1. Manual Decryption: Triggered if Options specify a cipher.
	if opts.cipher {
		decryptedValue, err := security.Decrypt(cast.ToString(value), security.WithCipherName(opts.cipherName))
		if err != nil {
			logger.Error("decrypt fail", "value", value, "cipher", opts.cipherName, "err", err.Error())
		} else {
			value = decryptedValue
		}
	}
	valueStr, ok := value.(string)
	if ok {
		// 2. Dynamic Parameter Rendering: Recursively resolves ${placeholder} within values.
		if configParamCompile.MatchString(valueStr) {
			for _, matchKey := range configParamCompile.FindAllString(valueStr, -1) {
				k1 := strings.Split(matchKey, "${")
				if len(k1) == 2 {
					k2 := strings.Split(k1[1], "}")
					if len(k2) == 2 {
						valueStr = strings.Replace(valueStr, matchKey, GetString(strings.TrimSpace(k2[0]), ""), -1)
					}
				}
			}
		}
		// 3. Auto-Decryption: Identifies 'encrypted_' prefix and resolves automatically using the internal security module.
		if !opts.disableAutoDecryptValue && strings.HasPrefix(valueStr, ValueEncryptFlag) {
			kv := strings.Split(valueStr, ValueEncryptSplitSymbol)
			cn := ""
			if len(kv) > 1 {
				cv := strings.Split(kv[0], ValueEncryptCipherNameSplitSymbol)
				if len(cv) > 1 {
					cn = cv[1]
				}
			}
			if cn != "" {
				ev := strings.Join(kv[1:], ValueEncryptSplitSymbol)
				dv, err := security.Decrypt(ev, security.WithCipherName(cn))
				if err != nil {
					logger.Error("decrypt fail", value, ev, "cipher", cn, "err", err)
				} else {
					valueStr = dv
				}
			}
		}

		value = valueStr
	}
	return value
}

// setValue processes a value (including encryption) and pushes it to target sources.
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

// setValueToSource pushes configuration to a specific source or all sources if sourceName is empty.
func (m *ConfigManager) setValueToSource(key, sourceName string, value any) error {
	if sourceName == "" {
		m.sm.RLock()
		for _, sourcer := range m.sourcers {
			logger.Debug("set key to source", "key", key, "source", sourcer.Name(), "value", value)
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

// disconnect triggers shutdown for all registered configuration sources.
func (m *ConfigManager) disconnect() {
	m.sm.RLock()
	defer m.sm.RUnlock()
	for name, sourcer := range m.sourcers {
		logger.Debug("stop config source", "source", name)
		sourcer.Disconnect()
	}
}

// Set adds or updates a configuration in local memory or a remote config center.
func Set(key string, value any, opts ...Option) error {
	return configmanager.setValue(key, value, GetOptions(opts...))
}

// Get retrieves a configuration with automatic decryption and parameter resolution.
func Get(key string, options *Options) any {
	return configmanager.getValue(key, options)
}

// GetWithPrefix retrieves configurations by prefix. Keys in the resulting map are properties-formatted.
func GetWithPrefix(prefixKey string, opts ...Option) map[string]any {
	return configmanager.getValueByPrefix(prefixKey, GetOptions(opts...))
}

// GetString retrieves a string value with a fallback default. Supports case transformation via Options.
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

// GetStrings retrieves a slice of strings using a configurable delimiter (default is ',').
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

// GetByte retrieves a value as a byte slice.
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
	return []byte(value)
}

// GetBool converts various representations (true, 1, yes, on) to a boolean.
func GetBool(key string, defaultValue bool, opts ...Option) bool {
	v := Get(key, GetOptions(opts...))
	if v == nil {
		return defaultValue
	}
	value, _ := ccast.ToBoolE(v)
	return value
}

// GetBools retrieves a slice of booleans from a delimited string.
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

// GetInt retrieves a value and casts it to an integer.
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

// GetInts retrieves a slice of integers from a delimited string.
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

// GetInt64 retrieves a value as an int64.
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

// GetInt64s retrieves a slice of int64s.
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

// GetInt32 retrieves a value as an int32.
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

// GetInt32s retrieves a slice of int32s.
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

// GetFloat64 retrieves a value as a float64.
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

// GetFloat64s retrieves a slice of float64s.
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

// GetFloat32 retrieves a value as a float32.
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

// GetFloat32s retrieves a slice of float32s.
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

// GetDuration parses a string representation of time (e.g., "1h", "30s") into a time.Duration.
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

// GetTime parses a value into time.Time using the specified location in Options.
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

// Exist checks if a configuration key exists in the global store.
func Exist(key string) bool {
	options := GetOptions()
	return Get(key, options) != nil
}

// GetAndUnmarshal retrieves a configuration string and deserializes it into the target pointer.
func GetAndUnmarshal(key string, outPtr any, opts ...Option) error {
	options := GetOptions(opts...)
	if options.unmarshaler != nil {
		return options.unmarshaler.Unmarshal([]byte(GetString(key, "", opts...)), outPtr)
	}
	return GetAndJsonUnmarshal(key, outPtr, opts...)
}

// GetAndJsonUnmarshal retrieves a value and deserializes it using standard JSON rules.
func GetAndJsonUnmarshal(key string, outPtr any, opts ...Option) error {
	return json.Unmarshal([]byte(GetString(key, "", opts...)), outPtr)
}

// GetAndYamlUnmarshal retrieves a value and deserializes it using standard YAML rules.
func GetAndYamlUnmarshal(key string, outPtr any, opts ...Option) error {
	return yaml.Unmarshal([]byte(GetString(key, "", opts...)), outPtr)
}

// GetWithUnmarshal retrieves configurations by prefix and unmarshals the resulting object.
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

// GetWithJsonUnmarshal retrieves configurations by prefix and unmarshals using JSON.
func GetWithJsonUnmarshal(prefix string, outPtr any, opts ...Option) error {
	configMap := getConfigMap(GetWithPrefix(prefix, opts...))
	outBytes, err := json.Marshal(&configMap)
	if err != nil {
		return err
	}
	return json.Unmarshal(outBytes, outPtr)
}

// GetWithYamlUnmarshal retrieves configurations by prefix and unmarshals using YAML.
func GetWithYamlUnmarshal(prefix string, outPtr any, opts ...Option) error {
	configMap := getConfigMap(GetWithPrefix(prefix, opts...))
	outBytes, err := yaml.Marshal(&configMap)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(outBytes, outPtr)
}

// GetKeyValueToMap parses a single key-value pair into a structured map, supporting array notation.
func GetKeyValueToMap(key string, value any) map[string]any {
	return getConfigMap(map[string]any{
		key: value,
	})
}

// AddListener attaches a callback for updates to a specific direct key.
func AddListener(key string, callback func(*Event)) {
	options := GetOptions(WithWatch(callback))
	configmanager.addListener(key, options)
}

// AddPatternListener attaches a callback using regex pattern matching for keys.
func AddPatternListener(pattern string, callback func(*Event)) {
	options := GetOptions(WithMatchWatch(pattern, callback))
	configmanager.addListener("", options)
}

// AddPrefixListener attaches a callback for updates to any key starting with a specific prefix.
func AddPrefixListener(prefix string, callback func(*Event)) {
	options := GetOptions(WithPrefixWatch(prefix, callback))
	configmanager.addListener("", options)
}

// RemoveListener unregisters all callbacks for the specified key.
func RemoveListener(key string) {
	configmanager.removeListener(key)
}

// getConfigMap expands a flattened properties-style map (e.g., "a.b.c") into a nested map structure.
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

// mergeConfigMap deeply merges source map data into a destination map.
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

// getConfigValue recursively generates nested map structures, including special handling for list notation like key[0].
func getConfigValue(index int, keyList []string, value any, keyValue map[string]any, skipKeys map[string]struct{}) map[string]any {
	key := keyList[index]
	// Base case: reaching the leaf node of the key path.
	if index == len(keyList)-1 {
		if !strings.HasSuffix(key, "]") {
			return map[string]any{
				key: value,
			}
		}
		// List handling: collect all items with matching index notation.
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
	// Middle case: handle nested objects within lists.
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
	// Recursive case: step deeper into the key path.
	return map[string]any{
		key: getConfigValue(index+1, keyList, value, keyValue, skipKeys),
	}
}
