// Package api @Author larry
// @Date 2026/06/21
// @Desc 凭邀请码查邀请人信息（问题4，老人接受页展示「是否接受 XX 的邀请」）

package api

import (
	"time"

	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// InviterInfoReq 查邀请人信息请求（需登录；inviteCode 从 query 绑定）
type InviterInfoReq struct {
	session.Session
	InviteCode string `form:"inviteCode" validate:"required"`
}

// InviterInfo 凭邀请码查邀请人头像/昵称/脱敏电话/称呼（任意登录用户可查，邀请失效返错误码）
func (a *ApiInvitation) InviterInfo(req InviterInfoReq) (interface{}, error) {
	return service.GetInvitationService().InviterInfo(req.InviteCode, time.Now().UnixMilli())
}
