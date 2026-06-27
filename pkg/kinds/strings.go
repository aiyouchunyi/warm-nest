// Package kind @Author larry
// @Date 2024/4/17 09:58:00
// @Desc 自定义字符串数组
package kinds

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Strings []string

// Value implements the driver.Valuer interface
func (a Strings) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *Strings) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("Type assertion .([]byte) failed.")
	}

	return json.Unmarshal(bytes, a)
}

// Get 获取指定索引的数据，索引越界返回空字符串
func (a Strings) Get(index int) string {
	if index < 0 || index >= len(a) {
		return ""
	}
	return a[index]
}
