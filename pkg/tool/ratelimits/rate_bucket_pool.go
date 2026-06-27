// Package ratelimits @Author larry
// @Date 2025/4/1 14:45
// @Desc

package ratelimits

import (
	"time"

	"github.com/juju/ratelimit"

	"warm-nest/pkg/tool/concurrent"
)

var rateBucketPool = concurrent.NewCMap[*ratelimit.Bucket]()

// Wait 等待
func Wait(key string, rateInSec int) {
	bucket, _ := GetBucket(key, rateInSec)
	bucket.Wait(1)
}

// WaitDuration 等待最大时间
func WaitDuration(key string, rateInSec int, maxWait time.Duration) bool {
	bucket, _ := GetBucket(key, rateInSec)
	return bucket.WaitMaxDuration(1, maxWait)
}

// GetBucket 获取限流桶
func GetBucket(key string, rateInSec int) (*ratelimit.Bucket, bool) {
	return rateBucketPool.LoadOrGen(key, func() *ratelimit.Bucket {
		return ratelimit.NewBucket(time.Second/time.Duration(rateInSec), int64(rateInSec))
	})
}
