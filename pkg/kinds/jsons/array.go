// Package jsons @Author larry
// @Date 2024/4/17 09:58:00
// @Desc 自定义字符串数组
package jsons

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Array[V any] []V

// Value implements the driver.Valuer interface
func (a Array[V]) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *Array[V]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(bytes, a)
}

// Append 添加元素
func (a *Array[V]) Append(v V) {
	if *a == nil {
		*a = make(Array[V], 0)
	}
	*a = append(*a, v)
}
