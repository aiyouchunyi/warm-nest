// Package api @Author larry
// @Date 2026/06/18
// @Desc 查询当前用户绑定状态（PRD §8.0.3 未经邀请进入拦截）

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// BindStatusReq 绑定状态查询请求（需登录，无业务入参）
type BindStatusReq struct {
	session.Session
}

// BindStatus 查当前用户是否已绑定（供前端拦截未经邀请进入的老人到等待页）
func (a *ApiFamily) BindStatus(req BindStatusReq) (interface{}, error) {
	return service.GetFamilyService().GetBindStatus(req.ReqUser)
}
