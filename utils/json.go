package utils

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/spf13/cast"
)

/*
JSONDuration 字符串或者数字json反序列化为time.Duration格式
*/
type JSONDuration struct {
	time.Duration
}

func (d JSONDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *JSONDuration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		return err
	default:
		return errors.New("invalid duration")
	}
}

/*
JSONStrings 逗号分隔的字符串或者字符串列表序列化为[]string
[]string转换为逗号分隔的字符串
*/
type JSONStrings []string

func (s JSONStrings) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.Join(s, ",") + `"`), nil
}

// UnmarshalJSON 列表或者字符串反序列化为字符串列表
func (s *JSONStrings) UnmarshalJSON(b []byte) error {
	n := len(b)
	if n <= 2 {
		return nil
	}
	// 字符串
	if b[0] == '"' {
		*s = strings.Split(string(b[1:n-1]), ",")
		return nil
	} else if b[0] == '[' {
		// 列表
		var out []any
		if err := json.Unmarshal(b, &out); err != nil {
			return err
		}
		result := make([]string, 0, len(out))
		for _, v := range out {
			result = append(result, cast.ToString(v))
		}
		*s = result
		return nil
	}
	return errors.New("invliad strings")
}

const (
	// 删除标志
	DelFlag = "-"
	// 追加标志
	AppendFlag = "+"
	// 替换标志
	ReplaceFlag = "="
	// 分隔符
	SplitFlag = ":"
)

// Merge 两个列表合并
// 系统中有很多内建的配置,如果需要修改则需要修改整个列表
// 通过此方法可以将内建配置和用户配置进行合并
// -a => 删除a
// +a:b => a后面追加b
// =a:b => a替换为b
// 保持顺序不变
// a = ["a", "b", "c"]
// b = ["-a", "+b:b1", "=d:dd", "e"]
// 合并后
// c = ["b","b1", "c", "dd", "e"]
func (s JSONStrings) Merge(cs JSONStrings) JSONStrings {
	var ns JSONStrings
	for _, v := range s {
		values := JSONStrings{v}
		for _, v1 := range cs {
			if v1 == DelFlag+v {
				values = JSONStrings{}
				continue
			}
			if strings.HasPrefix(v1, AppendFlag+v+SplitFlag) {
				values = JSONStrings{v, strings.TrimPrefix(v1, AppendFlag+v+SplitFlag)}
				continue
			}
			if strings.HasPrefix(v1, ReplaceFlag+v+SplitFlag) {
				values = JSONStrings{strings.TrimPrefix(v1, ReplaceFlag+v+SplitFlag)}
				continue
			}
		}
		if len(values) != 0 {
			for _, v := range values {
				exist := false
				for _, v1 := range ns {
					if v1 == v {
						exist = true
						break
					}
				}
				if !exist {
					ns = append(ns, v)
				}
			}
		}
	}
	for _, v := range cs {
		if strings.HasPrefix(v, DelFlag) ||
			strings.HasPrefix(v, AppendFlag) ||
			strings.HasPrefix(v, ReplaceFlag) {
			continue
		}
		exist := false
		for _, v1 := range ns {
			if v1 == v {
				exist = true
				break
			}
		}
		if !exist {
			ns = append(ns, v)
		}
	}
	return ns
}
