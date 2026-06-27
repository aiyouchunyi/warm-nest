// Package api @Author larry
// @Date 2026/06/22
// @Desc 凭 phoneCode 换手机号明文（问题5：拆分手机号采集与邀请创建）
//
// 设计：前端 <button open-type="getPhoneNumber"> 拿到一次性 code 后，先调本接口换成
// 明文手机号回填到表单，用户确认后再随 /invitation/create 或 /user/profile 提交明文 phone。
// 即「先有接口按 code 拿手机号，再直接传手机号」，create/profile 不再直收 code。
package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// ResolvePhoneReq 凭 phoneCode 换手机号请求（需登录）
type ResolvePhoneReq struct {
	session.Session
	PhoneCode string `json:"phoneCode" validate:"required"` // getPhoneNumber 回调的动态令牌（5min 有效、一次性）
}

// ResolvePhoneResult 换号结果
type ResolvePhoneResult struct {
	Phone string `json:"phone"` // 明文手机号（无区号），供前端回填后再提交
}

// ResolvePhone 用 phoneCode 换真实手机号明文返回（不落库，仅换号）。⚠️ 手机号敏感，勿打日志。
func (a *ApiUser) ResolvePhone(req ResolvePhoneReq) (interface{}, error) {
	phone, err := service.GetUserService().ResolvePhone(req.PhoneCode)
	if err != nil {
		return nil, err
	}
	return &ResolvePhoneResult{Phone: phone}, nil
}
