// Package api @Author larry
// @Date 2026/06/15
// @Desc 查询发起的守护邀请列表

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// ListInvitationReq 查询邀请列表请求（子女端，需登录）
type ListInvitationReq struct {
	session.Session
}

// ListInvitation 子女查自己发起的邀请及状态
func (a *ApiInvitation) ListInvitation(req ListInvitationReq) (interface{}, error) {
	return service.GetInvitationService().ListByGuardian(req.ReqUser)
}
