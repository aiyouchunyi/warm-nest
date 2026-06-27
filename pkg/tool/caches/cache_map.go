// Package caches @Author larry
// @Date 2025/7/25 13:42
// @Desc

package caches

import (
	"warm-nest/pkg/utils/maps"
	"warm-nest/pkg/utils/reflects"
)

type LoadItemFunc func(itemKey string) (interface{}, error)

type CacheMap[T any] struct {
	cache        *Cache
	cacheKey     string
	ttlInSec     int64
	loadFunc     LoadFunc
	loadItemFunc LoadItemFunc
}

func NewCacheMap[T any](cacheKey string, ttlInSec int64, loadFunc LoadFunc, loadItemFunc LoadItemFunc) *CacheMap[T] {
	return &CacheMap[T]{
		cache:        NewCache(),
		cacheKey:     cacheKey,
		ttlInSec:     ttlInSec,
		loadFunc:     loadFunc,
		loadItemFunc: loadItemFunc,
	}
}

// GetItem 获取单个缓存项，如果不存在则加载
func (cm *CacheMap[T]) GetItem(itemKey string) (T, bool) {
	mapValue, err := cm.Get()
	var zeroValue T
	if err != nil {
		return zeroValue, false
	}
	value, ok := mapValue[itemKey]
	return value, ok
}

// RefreshItem 刷新单个缓存项
func (cm *CacheMap[T]) RefreshItem(itemKey string) error {
	mapValue, err := cm.Get()
	if err != nil {
		return err
	}
	// 重新加载该项
	value, err := cm.loadItemFunc(itemKey)
	if err != nil {
		return err
	}
	if value == nil {
		delete(mapValue, itemKey)
	} else {
		mapValue[itemKey] = value.(T)
	}
	cm.cache.Set(cm.cacheKey, mapValue, cm.ttlInSec)
	return nil
}

// SetItem 设置或更新单个缓存项
func (cm *CacheMap[T]) SetItem(itemKey string, value T) bool {
	mapValue, err := cm.Get()
	if err != nil {
		return false
	}
	if reflects.IsNil(value) {
		delete(mapValue, itemKey)
	} else {
		mapValue[itemKey] = value
	}
	cm.cache.Set(cm.cacheKey, mapValue, cm.ttlInSec)
	return true
}

// DeleteItem 删除单个缓存项
func (cm *CacheMap[T]) DeleteItem(itemKey string) bool {
	mapValue, err := cm.Get()
	if err != nil {
		return false
	}
	delete(mapValue, itemKey)
	cm.cache.Set(cm.cacheKey, mapValue, cm.ttlInSec)
	return true
}

// Get 获取整个缓存映射
func (cm *CacheMap[T]) Get() (map[string]T, error) {
	mapValue, err := cm.cache.GetOrLoad(cm.cacheKey, cm.ttlInSec, cm.loadFunc)
	if err != nil {
		return nil, err
	}
	return mapValue.(map[string]T), nil
}

// GetAll 获取所有缓存项
func (cm *CacheMap[T]) GetAll() ([]T, error) {
	mapValue, err := cm.Get()
	if err != nil {
		return nil, err
	}
	return maps.Values(mapValue), nil
}

func (cm *CacheMap[T]) ClearCache() {
	cm.cache.Delete(cm.cacheKey)
}
