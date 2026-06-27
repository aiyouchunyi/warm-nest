// Package api @Author larry
// @Date 2026/06/19
// @Desc 撤销守护邀请

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// CancelInvitationReq 撤销邀请请求（子女端，需登录）
type CancelInvitationReq struct {
	session.Session
	InvitationId string `json:"invitationId" validate:"required"`
}

// CancelInvitation 子女撤销自己发起的待接受邀请（PENDING→CANCELLED）
func (a *ApiInvitation) CancelInvitation(req CancelInvitationReq) (interface{}, error) {
	if err := service.GetInvitationService().Cancel(req.ReqUser, req.InvitationId); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
