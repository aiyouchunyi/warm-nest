// Package concurrent @Author larry
// @Date 2025/5/13 15:46
// @Desc

package concurrent

import (
	"sync"

	"golang.org/x/sync/singleflight"
)

// CMap 并发安全的 Map 工具类
type CMap[V any] struct {
	data   sync.Map
	flight singleflight.Group
}

// NewCMap 创建一个新的 CMap 实例
func NewCMap[V any]() *CMap[V] {
	return &CMap[V]{}
}

// Store 设置键值对
func (m *CMap[V]) Store(key string, value V) {
	m.data.Store(key, value)
}

func (m *CMap[V]) StoreIfAbsent(key string, value V) bool {
	_, ok := m.data.Load(key)
	if ok {
		return false
	}
	m.data.Store(key, value)
	return true
}

// LoadOrStore 设置键值对
func (m *CMap[V]) LoadOrStore(key string, value V) (V, bool) {
	v, ok := m.data.LoadOrStore(key, value)
	return v.(V), ok
}

// LoadOrGen 设置键值对
func (m *CMap[V]) LoadOrGen(key string, loader func() V) (V, bool) {
	value, ok := m.Load(key)
	if ok {
		return value, ok
	}
	v, _, _ := m.flight.Do(key, func() (interface{}, error) {
		return loader(), nil
	})
	return m.LoadOrStore(key, v.(V))
}

func (m *CMap[V]) Set(key string, loader func() V) bool {
	v, _, _ := m.flight.Do(key, func() (interface{}, error) {
		return loader(), nil
	})
	m.Store(key, v.(V))
	return true
}

func (m *CMap[V]) SetIfAbsent(key string, loader func() V) bool {
	_, ok := m.Load(key)
	if ok {
		return false
	}
	v, _, _ := m.flight.Do(key, func() (interface{}, error) {
		return loader(), nil
	})
	m.StoreIfAbsent(key, v.(V))
	return true
}

// Load 获取键对应的值
func (m *CMap[V]) Load(key string) (value V, ok bool) {
	v, ok := m.data.Load(key)
	if ok {
		value = v.(V)
	}
	return
}

// Delete 删除键值对
func (m *CMap[V]) Delete(key string) {
	m.data.Delete(key)
}

// Has 检查键是否存在
func (m *CMap[V]) Has(key string) bool {
	_, ok := m.data.Load(key)
	return ok
}
