// Package jsons @Author larry
// @Date 2024/4/17 09:58:00
// @Desc 自定义字符串数组
package jsons

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type JSONMap map[string]JSONObject

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(bytes, j)
}

func NewJSONMap() JSONMap {
	return make(JSONMap)
}

// GetVal 获取指定key的值
func (j *JSONMap) GetVal(key string) JSONObject {
	if j == nil {
		return NewJSON()
	}
	return (*j)[key]
}
