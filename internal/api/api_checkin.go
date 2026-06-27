// Package api @Author larry
// @Date 2026/06/15
// @Desc 打卡接口聚合（struct + singleton）

package api

import "sync"

// ApiCheckIn 打卡接口
type ApiCheckIn struct{}

var apiCheckIn *ApiCheckIn
var apiCheckInOnce sync.Once

// GetApiCheckIn 获取打卡接口单例
func GetApiCheckIn() *ApiCheckIn {
	apiCheckInOnce.Do(func() {
		apiCheckIn = &ApiCheckIn{}
	})
	return apiCheckIn
}
