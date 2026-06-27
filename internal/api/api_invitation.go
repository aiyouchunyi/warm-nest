// Package api @Author larry
// @Date 2026/06/15
// @Desc 守护邀请接口聚合（struct + singleton）

package api

import "sync"

// ApiInvitation 守护邀请接口
type ApiInvitation struct{}

var apiInvitation *ApiInvitation
var apiInvitationOnce sync.Once

// GetApiInvitation 获取守护邀请接口单例
func GetApiInvitation() *ApiInvitation {
	apiInvitationOnce.Do(func() {
		apiInvitation = &ApiInvitation{}
	})
	return apiInvitation
}
