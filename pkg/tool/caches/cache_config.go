// Package caches @Author larry
// @Date 2025/4/25 11:41
// @Desc

package caches

type CacheConfig struct {
	Renewal  bool  `gorm:"comment:是否续期"`
	Log      bool  `gorm:"comment:是否打印日志"`
	TtlInSec int64 `gorm:"comment:过期时间s"`
}

type Option func(c *CacheConfig)

func NewCacheConfig(opts ...Option) *CacheConfig {
	config := &CacheConfig{
		Renewal:  false,
		Log:      false,
		TtlInSec: 0,
	}
	for _, opt := range opts {
		opt(config)
	}
	return config
}

func Renewal(renewal bool) Option {
	return func(c *CacheConfig) {
		c.Renewal = renewal
	}
}

func TtlInSec(ttlInSec int64) Option {
	return func(c *CacheConfig) {
		c.TtlInSec = ttlInSec
	}
}

func Log(log bool) Option {
	return func(c *CacheConfig) {
		c.Log = log
	}
}
