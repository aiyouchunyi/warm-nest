// Package kind @Author larry
// @Date 2024/4/17 09:58:00
// @Desc 自定义字符串数组
package kinds

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"warm-nest/pkg/utils/maps"
)

type KKMap[K comparable, V any] map[K]KMap[K, V]

// Value implements the driver.Valuer interface
func (m KKMap[K, V]) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface
func (m *KKMap[K, V]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	return json.Unmarshal(bytes, m)
}

// Set 安全设置值
func (m *KKMap[K, V]) Set(k1, k2 K, value V) {
	if *m == nil {
		*m = make(KKMap[K, V])
	}
	if _, ok := (*m)[k1]; !ok {
		(*m)[k1] = make(map[K]V)
	}
	(*m)[k1][k2] = value
}

// Get safely gets a value
func (m KKMap[K, V]) Get(k1, k2 K) (V, bool) {
	vMap, ok := m[k1]
	if !ok || vMap == nil {
		var v V
		return v, false
	}
	v, ok := vMap[k2]
	return v, ok
}

// GetMap safely gets a value
func (m KKMap[K, V]) GetMap(k1 K) KMap[K, V] {
	vMap, ok := m[k1]
	if !ok || vMap == nil {
		return make(map[K]V)
	}
	return vMap
}

// SetMap safely sets a value
func (m *KKMap[K, V]) SetMap(k1 K, vMap map[K]V) {
	if *m == nil {
		*m = make(KKMap[K, V])
	}
	(*m)[k1] = vMap
}

func (m KKMap[K, V]) Keys() []K {
	return maps.Keys(m)
}

func (m KKMap[K, V]) SecondKeys(k1 K) []K {
	return maps.Keys(m[k1])
}
