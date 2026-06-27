// Package api @Author larry
// @Date 2026/06/15
// @Desc 用户登录接口

package api

import (
	"warm-nest/internal/service"
)

// LoginReq 登录请求
type LoginReq struct {
	Code string `json:"code" validate:"required"` // 小程序 wx.login 的 code
}

// Login 小程序登录（免登录接口）
func (a *ApiUser) Login(req LoginReq) (interface{}, error) {
	return service.GetUserService().LoginByCode(req.Code)
}
