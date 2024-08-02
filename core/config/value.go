package config

import "fmt"

// Value 配置值
type Value struct {
	// 值所在的配置源
	Sourcer Sourcer
	// 值
	Value any
	// 值的引用
	// 例如删除一个文件，是不知道删除的key
	// 此时可以使用此字段，删除引用下的所有key
	Ref string
	// 值的优先级
	// 同一个配置源中也会出现同一个配置有不同的优先级
	Priority int
}

// String 字符串格式化
func (v Value) String() string {
	if v.Sourcer != nil {
		return fmt.Sprintf("sourcer: '%s', value: '%+v', ref: '%s'",
			v.Sourcer.Name(), v.Value, v.Ref)
	}
	return fmt.Sprintf("sourcer: 'nil', value: '%+v', ref: '%s'",
		v.Value, v.Ref)
}
