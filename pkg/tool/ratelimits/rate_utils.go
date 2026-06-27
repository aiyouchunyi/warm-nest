// Package ratelimits @Author larry
// @Date 2025/4/1 14:52
// @Desc

package ratelimits

import (
	"warm-nest/pkg/utils/strings"
)

// RateKey 生成限流键
func RateKey(prefix string, params ...string) string {
	return strings.Join(strings.ColonSplitChar, "R-", prefix, strings.JoinV2(strings.UnderlineSplitStr, params))
}
