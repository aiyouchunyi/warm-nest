// Package api @Author larry
// @Date 2026/06/15
// @Desc 用户接口聚合（struct + singleton，具体方法各自独立文件）

package api

import "sync"

// ApiUser 用户接口
type ApiUser struct{}

var apiUser *ApiUser
var apiUserOnce sync.Once

// GetApiUser 获取用户接口单例
func GetApiUser() *ApiUser {
	apiUserOnce.Do(func() {
		apiUser = &ApiUser{}
	})
	return apiUser
}
