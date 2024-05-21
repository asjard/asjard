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
}

// String 字符串格式化
func (v Value) String() string {
	return fmt.Sprintf("sourcer: '%s', value: '%+v', ref: '%s'",
		v.Sourcer.Name(), v.Value, v.Ref)
}
