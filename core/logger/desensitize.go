package logger

import (
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

const defaultMaxDesensitizeDepth = 8

// DesensitizeKey is a sensitive field / header name and/or content pattern,
// with optional keep / replacement overrides.
type DesensitizeKey struct {
	ReplacementConfig
	Name    string `json:"name"`
	Pattern string `json:"pattern"`
}

// DesensitizeConfig controls global log field desensitization.
// Disabled defaults to false (desensitization on).
// Mask format: keepPrefix + replacement + keepSuffix; if length <= prefix+suffix, use replacement only.
type DesensitizeConfig struct {
	ReplacementConfig
	Disabled    bool             `json:"disabled"`
	Keys        []DesensitizeKey `json:"keys"`
	BuiltinKeys []DesensitizeKey `json:"builtInKeys"`
	MaxDepth    int              `json:"maxDepth"`
}

// ReplacementConfig defines mask placeholder and how many runes to keep on each side.
type ReplacementConfig struct {
	KeepPrefix  int    `json:"keepPrefix"`
	KeepSuffix  int    `json:"keepSuffix"`
	Replacement string `json:"replacement"`
}

type replacementRule struct {
	replacement string
	reg         *regexp.Regexp
	keepPrefix  int
	keepSuffix  int
}

// Desensitizer holds compiled desensitization rules.
type Desensitizer struct {
	keys     map[string]replacementRule
	patterns []replacementRule
	maxDepth int
}

// NewDesensitizer builds a Desensitizer from config.
// Returns nil when desensitization is disabled.
//
//gocyclo:ignore
func NewDesensitizer(cfg *DesensitizeConfig) *Desensitizer {
	if cfg == nil || cfg.Disabled {
		return nil
	}

	maxDepth := cfg.MaxDepth
	if maxDepth <= 0 {
		maxDepth = defaultMaxDesensitizeDepth
	}

	keyRules := make(map[string]replacementRule, len(cfg.BuiltinKeys)+len(cfg.Keys))
	patternRules := make([]replacementRule, 0, len(cfg.BuiltinKeys)+len(cfg.Keys))
	patternMap := make(map[string]struct{}, len(cfg.BuiltinKeys)+len(cfg.Keys))

	// User keys first so they take precedence over builtin keys/patterns.
	for _, k := range append(cfg.Keys, cfg.BuiltinKeys...) {
		if k.Name == "" && k.Pattern == "" {
			continue
		}
		rule := replacementRule{
			replacement: k.Replacement,
			keepPrefix:  k.KeepPrefix,
			keepSuffix:  k.KeepSuffix,
		}
		if rule.replacement == "" {
			rule.replacement = cfg.Replacement
		}
		if rule.keepPrefix == 0 && cfg.KeepPrefix > 0 {
			rule.keepPrefix = cfg.KeepPrefix
		}
		if rule.keepSuffix == 0 && cfg.KeepSuffix > 0 {
			rule.keepSuffix = cfg.KeepSuffix
		}
		if k.Pattern != "" {
			if _, ok := patternMap[k.Pattern]; ok {
				continue
			}
			reg, err := regexp.Compile(k.Pattern)
			if err != nil {
				continue
			}
			rule.reg = reg
			patternRules = append(patternRules, rule)
			patternMap[k.Pattern] = struct{}{}
			continue
		}
		name := strings.ToLower(strings.TrimSpace(k.Name))
		if name == "" {
			continue
		}
		if _, ok := keyRules[name]; !ok {
			keyRules[name] = rule
		}
	}

	if len(keyRules) == 0 && len(patternRules) == 0 {
		return nil
	}

	return &Desensitizer{
		keys:     keyRules,
		patterns: patternRules,
		maxDepth: maxDepth,
	}
}

func (r replacementRule) getKeepPrefix() int {
	if r.keepPrefix < 0 {
		return 0
	}
	return r.keepPrefix
}

func (r replacementRule) getKeepSuffix() int {
	if r.keepSuffix < 0 {
		return 0
	}
	return r.keepSuffix
}

// MaskAttr returns a desensitized copy of attr.
func (d *Desensitizer) MaskAttr(attr slog.Attr) slog.Attr {
	if d == nil {
		return attr
	}
	attr.Value = d.maskSlogValue(attr.Key, attr.Value, 0)
	return attr
}

// MaskValue desensitizes an arbitrary Go value under the given key.
func (d *Desensitizer) MaskValue(key string, v any) any {
	if d == nil {
		return v
	}
	return d.maskAny(key, v, 0)
}

func (d *Desensitizer) keyRule(key string) (replacementRule, bool) {
	if key == "" || len(d.keys) == 0 {
		return replacementRule{}, false
	}
	rule, ok := d.keys[strings.ToLower(key)]
	return rule, ok
}

func (d *Desensitizer) maskSlogValue(key string, v slog.Value, depth int) slog.Value {
	if depth > d.maxDepth {
		return v
	}
	switch v.Kind() {
	case slog.KindGroup:
		attrs := v.Group()
		out := make([]slog.Attr, len(attrs))
		for i, a := range attrs {
			out[i] = d.MaskAttr(a)
		}
		return slog.GroupValue(out...)
	case slog.KindLogValuer:
		return d.maskSlogValue(key, v.LogValuer().LogValue(), depth+1)
	case slog.KindString:
		return slog.StringValue(d.maskString(key, v.String()))
	case slog.KindAny:
		return slog.AnyValue(d.maskAny(key, v.Any(), depth))
	default:
		if rule, ok := d.keyRule(key); ok {
			return slog.StringValue(rule.replacement)
		}
		return v
	}
}

// maskText applies keepPrefix + replacement + keepSuffix.
// If rune length <= prefix+suffix (or both are 0), returns replacement.
func (d *Desensitizer) maskText(s string, rule replacementRule) string {
	if s == "" {
		return s
	}
	keepPrefix := rule.getKeepPrefix()
	keepSuffix := rule.getKeepSuffix()
	if keepPrefix <= 0 && keepSuffix <= 0 {
		return rule.replacement
	}
	r := []rune(s)
	if len(r) <= keepPrefix+keepSuffix {
		return rule.replacement
	}
	return string(r[:keepPrefix]) + rule.replacement + string(r[len(r)-keepSuffix:])
}

func (d *Desensitizer) maskString(key string, val string) string {
	if val == "" {
		return val
	}
	if rule, ok := d.keyRule(key); ok {
		return d.maskText(val, rule)
	}
	for _, rule := range d.patterns {
		val = rule.reg.ReplaceAllStringFunc(val, func(m string) string {
			return d.maskText(m, rule)
		})
	}
	return val
}

//gocyclo:ignore
func (d *Desensitizer) maskAny(key string, v any, depth int) any {
	if v == nil || depth > d.maxDepth {
		return v
	}
	if rule, ok := d.keyRule(key); ok {
		return d.maskSensitiveValue(v, rule)
	}

	switch x := v.(type) {
	case string:
		return d.maskString(key, x)
	case []byte:
		return []byte(d.maskString(key, string(x)))
	case http.Header:
		return d.maskHeader(x)
	case slog.LogValuer:
		return d.maskSlogValue(key, x.LogValue(), depth+1).Any()
	case map[string]string:
		out := make(map[string]string, len(x))
		for k, val := range x {
			out[k] = d.maskString(k, val)
		}
		return out
	case map[string][]string:
		out := make(map[string][]string, len(x))
		for k, val := range x {
			out[k] = d.maskStringSlice(k, val)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = d.maskAny(k, val, depth+1)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = d.maskAny(key, val, depth+1)
		}
		return out
	case []string:
		return d.maskStringSlice(key, x)
	}

	rv := reflect.ValueOf(v)
	return d.maskReflect(key, rv, depth)
}

// maskSensitiveValue replaces a sensitive value while preserving a compatible type
// so reflect map/slice writes do not panic (e.g. map[string][]string).
func (d *Desensitizer) maskSensitiveValue(v any, rule replacementRule) any {
	if v == nil {
		return rule.replacement
	}
	switch x := v.(type) {
	case string:
		return d.maskText(x, rule)
	case []byte:
		return []byte(d.maskText(string(x), rule))
	case []string:
		if len(x) == 0 {
			return []string{rule.replacement}
		}
		out := make([]string, len(x))
		for i, val := range x {
			out[i] = d.maskText(val, rule)
		}
		return out
	case []any:
		out := make([]any, len(x))
		if len(out) == 0 {
			return []any{rule.replacement}
		}
		for i, val := range x {
			out[i] = d.maskSensitiveValue(val, rule)
		}
		return out
	default:
		return d.replacementFor(reflect.TypeOf(v), reflect.ValueOf(v), rule)
	}
}

func (d *Desensitizer) maskStringSlice(key string, vals []string) []string {
	if len(vals) == 0 {
		return vals
	}
	out := make([]string, len(vals))
	for i, val := range vals {
		out[i] = d.maskString(key, val)
	}
	return out
}

//gocyclo:ignore
func (d *Desensitizer) replacementFor(rt reflect.Type, rv reflect.Value, rule replacementRule) any {
	if rt == nil {
		return rule.replacement
	}
	switch rt.Kind() {
	case reflect.String:
		if rv.IsValid() && rv.Kind() == reflect.String {
			return d.maskText(rv.String(), rule)
		}
		return rule.replacement
	case reflect.Slice:
		if rt.Elem().Kind() == reflect.String {
			n := 1
			if rv.IsValid() && rv.Kind() == reflect.Slice {
				if rv.Len() == 0 {
					return []string{rule.replacement}
				}
				n = rv.Len()
			}
			out := reflect.MakeSlice(rt, n, n)
			for i := 0; i < n; i++ {
				val := rule.replacement
				if rv.IsValid() && rv.Kind() == reflect.Slice && i < rv.Len() {
					val = d.maskText(rv.Index(i).String(), rule)
				}
				out.Index(i).SetString(val)
			}
			return out.Interface()
		}
	case reflect.Array:
		if rt.Elem().Kind() == reflect.String {
			out := reflect.New(rt).Elem()
			for i := 0; i < rt.Len(); i++ {
				val := rule.replacement
				if rv.IsValid() && rv.Kind() == reflect.Array {
					val = d.maskText(rv.Index(i).String(), rule)
				}
				out.Index(i).SetString(val)
			}
			return out.Interface()
		}
	}
	return rule.replacement
}

func (d *Desensitizer) maskHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	out := make(http.Header, len(h))
	for k, vals := range h {
		out[k] = d.maskStringSlice(k, vals)
	}
	return out
}

//gocyclo:ignore
func (d *Desensitizer) maskReflect(key string, rv reflect.Value, depth int) any {
	if !rv.IsValid() || depth > d.maxDepth {
		return rv.Interface()
	}

	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.String:
		return d.maskString(key, rv.String())
	case reflect.Map:
		if rv.Type().Key().Kind() != reflect.String {
			return rv.Interface()
		}
		elemType := rv.Type().Elem()
		out := reflect.MakeMapWithSize(rv.Type(), rv.Len())
		for _, mk := range rv.MapKeys() {
			mv := rv.MapIndex(mk)
			masked := d.maskAny(mk.String(), mv.Interface(), depth+1)
			mvOut := reflect.ValueOf(masked)
			if !mvOut.IsValid() {
				continue
			}
			if !mvOut.Type().AssignableTo(elemType) {
				rule, ok := d.keyRule(mk.String())
				if !ok {
					continue
				}
				mvOut = reflect.ValueOf(d.replacementFor(elemType, mv, rule))
				if !mvOut.IsValid() || !mvOut.Type().AssignableTo(elemType) {
					continue
				}
			}
			out.SetMapIndex(mk, mvOut)
		}
		return out.Interface()
	case reflect.Slice, reflect.Array:
		if rule, ok := d.keyRule(key); ok {
			return d.replacementFor(rv.Type(), rv, rule)
		}
		n := rv.Len()
		out := make([]any, n)
		for i := 0; i < n; i++ {
			out[i] = d.maskAny(key, rv.Index(i).Interface(), depth+1)
		}
		return out
	case reflect.Struct:
		return d.maskStruct(rv, depth)
	default:
		return rv.Interface()
	}
}

func (d *Desensitizer) maskStruct(rv reflect.Value, depth int) any {
	rt := rv.Type()
	out := make(map[string]any, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if sf.PkgPath != "" { // unexported
			continue
		}
		name := sf.Name
		if tag := sf.Tag.Get("json"); tag != "" && tag != "-" {
			if comma := strings.IndexByte(tag, ','); comma >= 0 {
				tag = tag[:comma]
			}
			if tag != "" {
				name = tag
			}
		}
		fv := rv.Field(i)
		if !fv.CanInterface() {
			continue
		}
		out[name] = d.maskAny(name, fv.Interface(), depth+1)
	}
	return out
}
