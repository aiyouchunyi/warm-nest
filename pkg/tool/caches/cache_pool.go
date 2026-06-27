// Package caches @Author larry
// @Date 2025/4/25 10:31
// @Desc

package caches

import (
	"sync"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/reflects"
)

var cachePool *Cache
var cachePoolOnce sync.Once

func CachePool() *Cache {
	cachePoolOnce.Do(func() {
		cachePool = NewCache()
	})
	return cachePool
}

func GetOrLoad[V any](cacheKey string, ttlInSec int64, loader LoadFunc, opts ...Option) (V, error) {
	ck := NewCacheConfig(opts...)
	if ck.Log {
		logrus.WithFields(logrus.Fields{
			"cacheKey": cacheKey,
		}).Info("Cache GetOrLoad")
	}
	v, err := reflects.CastWithErr[V](CachePool().GetOrLoad(cacheKey, ttlInSec, loader))
	if err != nil {
		return v, err
	}
	if ck.Renewal {
		Set(cacheKey, v, ttlInSec)
	}
	return v, nil
}
func Get[V any](cacheKey string, opts ...Option) (V, bool) {
	ck := NewCacheConfig(opts...)
	if ck.Log {
		logrus.WithFields(logrus.Fields{
			"cacheKey": cacheKey,
		}).Info("Cache Get")
	}
	value, ok := CachePool().Get(cacheKey)
	v := reflects.MustCast[V](value)
	if ok && ck.Renewal && ck.TtlInSec > 0 {
		Set(cacheKey, v, ck.TtlInSec)
	}
	return v, ok
}

func Set(cacheKey string, value interface{}, ttlInSec int64, opts ...Option) {
	ck := NewCacheConfig(opts...)
	if ck.Log {
		logrus.WithFields(logrus.Fields{
			"cacheKey": cacheKey,
		}).Info("Cache Set")
	}
	CachePool().Set(cacheKey, value, ttlInSec)
}

func Delete(cacheKey string, opts ...Option) {
	ck := NewCacheConfig(opts...)
	if ck.Log {
		logrus.WithFields(logrus.Fields{
			"cacheKey": cacheKey,
		}).Info("Cache Delete")
	}
	CachePool().Delete(cacheKey)
}

func DeleteByPrefix(prefix string, opts ...Option) {
	ck := NewCacheConfig(opts...)
	if ck.Log {
		logrus.WithFields(logrus.Fields{
			"prefix": prefix,
		}).Info("Cache DeleteByPrefix")
	}
	CachePool().DeleteByPrefix(prefix)
}

func Keys() []string {
	return CachePool().Keys()
}
