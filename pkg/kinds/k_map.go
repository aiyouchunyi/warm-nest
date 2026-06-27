// Package kind @Author larry
// @Date 2024/4/17 09:58:00
// @Desc 自定义Map
package kinds

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"warm-nest/pkg/utils/maps"
)

type KMap[K comparable, V any] map[K]V

// Value implements the driver.Valuer interface
func (m KMap[K, V]) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface
func (m *KMap[K, V]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	return json.Unmarshal(bytes, m)
}

// Set 安全设置值
func (m *KMap[K, V]) Set(k K, v V) {
	if *m == nil {
		*m = make(KMap[K, V])
	}
	(*m)[k] = v
}

// Get 安全获取值
func (m KMap[K, V]) Get(k K) (V, bool) {
	v, ok := (m)[k]
	return v, ok
}

func (m KMap[K, V]) Keys() {
	maps.Keys(m)
}
