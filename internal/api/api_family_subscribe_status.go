// Package api @Author larry
// @Date 2026/06/18
// @Desc 查询当前用户服务号关注状态（PRD §8.2/§8.0.2.4 关注软引导）

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// SubscribeStatusReq 关注状态查询请求（需登录，无业务入参）
type SubscribeStatusReq struct {
	session.Session
}

// SubscribeStatusResp 关注状态
type SubscribeStatusResp struct {
	Subscribed bool `json:"subscribed"` // 当前用户是否已关注服务号
}

// SubscribeStatus 查当前用户是否已关注服务号（供前端决定是否展示关注软引导）
func (a *ApiFamily) SubscribeStatus(req SubscribeStatusReq) (interface{}, error) {
	subscribed, err := service.GetFanService().IsSubscribedByUserId(req.ReqUser)
	if err != nil {
		return nil, err
	}
	return SubscribeStatusResp{Subscribed: subscribed}, nil
}
