// Package jsons @Author larry
// File json_array.go
// @Date 2024/8/27 16:12:00
// @Desc
package jsons

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/transforms"
)

type JSONArray Array[JSONObject]

// Value implements the driver.Valuer interface
func (a JSONArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *JSONArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(bytes, a)
}

// NewArray 创建JSONArray
func NewArray() JSONArray {
	return make(JSONArray, 0)
}

// NewArrayFrom 从数据创建JSONArray
func NewArrayFrom(data any) JSONArray {
	if reflects.IsNil(data) {
		return nil
	}
	if reflects.IsString(data) {
		return NewArrayFromStr(data.(string))
	}
	return transforms.CopyTo[JSONArray](data)
}

// NewArrayFromStr 从字符串创建JSONArray
func NewArrayFromStr(data string) JSONArray {
	return transforms.Unmarshal[JSONArray](data)
}

// Append 添加元素
func (a *JSONArray) Append(v JSONObject) {
	if *a == nil {
		*a = make(JSONArray, 0)
	}
	*a = append(*a, v)
}

func (a *JSONArray) ToString() string {
	if a == nil {
		return "[]"
	}
	bytes, err := json.Marshal(a)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}
