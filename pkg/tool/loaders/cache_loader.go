// Package loaders @Author larry
// @Date 2025/9/16 19:30
// @Desc

package loaders

import (
	"sync"
	"time"
)

var GlobalCacheLoader = NewCacheLoader()

type cacheItem struct {
	value      interface{}
	expiration int64
}

type CacheLoader struct {
	data map[string]*cacheItem
	mu   sync.RWMutex
}

func NewCacheLoader() *CacheLoader {
	return &CacheLoader{
		data: make(map[string]*cacheItem),
	}
}

// Get 获取缓存，过期则通过 fetcher 获取新数据
func (c *CacheLoader) Get(key string, ttlInSec int64, fetcher func() (interface{}, error)) (interface{}, error) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	now := time.Now().UnixNano()
	if exists && item.expiration > now {
		return item.value, nil
	}

	// 过期或不存在，调用 fetcher
	val, err := fetcher()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.data[key] = &cacheItem{
		value:      val,
		expiration: now + ttlInSec*int64(time.Second),
	}
	c.mu.Unlock()
	return val, nil
}
