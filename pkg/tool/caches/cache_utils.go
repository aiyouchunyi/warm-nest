// Package caches @Author larry
// @Date 2024/12/24 13:43
// @Desc

package caches

import (
	"warm-nest/pkg/utils/strings"
)

// CacheKey 缓存键
func CacheKey(prefix string, params ...string) string {
	return strings.Join(strings.HyphenSplitStr, "C", prefix, strings.JoinV2(strings.UnderlineSplitStr, params))
}
