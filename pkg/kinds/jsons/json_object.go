// Package jsons @Author larry
// @Date 2024/4/17 09:58:00
// @Desc 自定义字符串数组
package jsons

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/creasty/defaults"

	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/transforms"
)

type JSONObject map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONObject) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONObject) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(bytes, j)
}

func NewJSON() JSONObject {
	return make(JSONObject)
}

// NewJSONForm 使用 JSON 编码实现结构体转 JSON
func NewJSONForm(src any) JSONObject {
	if reflects.IsNil(src) {
		return nil
	}
	if reflects.IsString(src) {
		return NewJSONFromStr(src.(string))
	}
	return transforms.CopyTo[JSONObject](src)
}

// NewJSONFromStr 使用 JSON 解码实现字符串转 JSON
func NewJSONFromStr(src string) JSONObject {
	return transforms.Unmarshal[JSONObject](src)
}

// ToStruct 使用 JSON 解码实现 JSON 转结构体
func ToStruct[T any](src JSONObject) T {
	return transforms.CopyTo[T](src)
}

func ToStructDefault[T any](src JSONObject) T {
	to := transforms.CopyTo[T](src)
	_ = defaults.Set(&to)
	return to
}

// Put 合并两个 JSON
func (j *JSONObject) Put(v JSONObject) {
	for key, value := range v {
		(*j)[key] = value
	}
}

// GetString 获取字符串
func (j JSONObject) GetString(key string) string {
	return reflects.MustCast[string](j[key])
}

// GetBool 获取布尔值
func (j JSONObject) GetBool(key string) bool {
	return reflects.MustCast[bool](j[key])
}

// SetVal 设置值
func (j JSONObject) SetVal(key string, value interface{}) JSONObject {
	j[key] = value
	return j
}

// Remove 删除键值对
func (j JSONObject) Remove(key string) JSONObject {
	delete(j, key)
	return j
}

// ToString 转换为json字符串
func (j JSONObject) ToString() string {
	return transforms.Marshal(j)
}
