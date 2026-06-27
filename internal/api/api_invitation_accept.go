// Package api @Author larry
// @Date 2026/06/15
// @Desc 接受守护邀请

package api

import (
	"time"

	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// AcceptInvitationReq 接受邀请请求（老人端，需先登录）
type AcceptInvitationReq struct {
	session.Session
	InviteCode string `json:"inviteCode" validate:"required"`
}

// AcceptInvitation 老人接受邀请，建立守护关系
func (a *ApiInvitation) AcceptInvitation(req AcceptInvitationReq) (interface{}, error) {
	err := service.GetInvitationService().Accept(req.ReqUser, req.InviteCode, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
