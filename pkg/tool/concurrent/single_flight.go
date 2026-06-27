// Package single @Author larry
// @Date 2025/3/14 15:05
// @Desc

package concurrent

import (
	"golang.org/x/sync/singleflight"

	"warm-nest/pkg/utils/strings"
)

var flight singleflight.Group

// SingleKey 单个键
func SingleKey(prefix string, params ...string) string {
	return strings.Join(strings.ColonSplitChar, "S-", prefix, strings.JoinV2(strings.UnderlineSplitStr, params))
}

// Do 执行单次操作
func Do(singleKey string, fn func() (interface{}, error)) (interface{}, error) {
	v, err, _ := flight.Do(singleKey, func() (interface{}, error) {
		return fn()
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}
