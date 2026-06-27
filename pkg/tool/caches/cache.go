// Package caches @Author larry
// @Date 2024/12/18 17:25
// @Desc

package caches

import (
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"

	"warm-nest/pkg/utils/maps"
	"warm-nest/pkg/utils/slices"
)

type LoadFunc func() (interface{}, error)

type Cache struct {
	c      *cache.Cache
	flight singleflight.Group
}

func NewCache() *Cache {
	return &Cache{
		c: cache.New(1*time.Minute, 1*time.Minute),
	}
}

func (cm *Cache) GetOrLoad(cacheKey string, ttlInSec int64, loader LoadFunc) (interface{}, error) {
	if v, ok := cm.Get(cacheKey); ok {
		return v, nil
	}

	v, err, _ := cm.flight.Do(cacheKey, func() (interface{}, error) {
		v, err := loader()
		if err == nil {
			cm.Set(cacheKey, v, ttlInSec)
		}
		return v, err
	})
	return v, err
}

func (cm *Cache) Get(cacheKey string) (interface{}, bool) {
	return cm.c.Get(cacheKey)
}

func (cm *Cache) Set(cacheKey string, value interface{}, ttlInSec int64) {
	cm.c.Set(cacheKey, value, time.Duration(ttlInSec)*time.Second)
}

func (cm *Cache) Delete(cacheKey string) {
	logrus.WithFields(logrus.Fields{
		"cacheKey": cacheKey,
	}).Info("Deleting cache key")
	cm.c.Delete(cacheKey)
}

func (cm *Cache) DeleteByPrefix(prefix string) {
	keys := make([]string, 0)
	for key := range cm.c.Items() {
		if strings.HasPrefix(key, CacheKey(prefix)) {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return
	}
	logrus.WithFields(logrus.Fields{
		"prefix": prefix,
		"keys":   keys,
	}).Info("Delete cache by prefix")
	for _, key := range keys {
		cm.c.Delete(key)
	}
}

func (cm *Cache) Keys() []string {
	keys := maps.Keys(cm.c.Items())
	return slices.Sort(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
}
