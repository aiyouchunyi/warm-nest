// Package api @Author larry
// @Date 2026/06/22
// @Desc 当前登录用户本人资料查看/编辑（问题2：子女查看编辑头像/昵称/手机号）
package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// MyProfileReq 查本人资料请求（需登录，无业务参数）
type MyProfileReq struct {
	session.Session
}

// MyProfile 查当前登录用户本人资料（问题2，供「我的资料」页回填头像/昵称/手机号）
func (a *ApiUser) MyProfile(req MyProfileReq) (interface{}, error) {
	return service.GetUserService().GetMyProfile(req.ReqUser)
}

// UpdateMyProfileReq 编辑本人资料请求（需登录）。
// 指针字段=patch 语义：不传该字段则不动（避免前端漏传把头像/手机号清空），传了（含空串）则覆盖。
type UpdateMyProfileReq struct {
	session.Session
	Nickname *string `json:"nickname"` // 微信昵称
	Avatar   *string `json:"avatar"`   // 头像 URL（wx.chooseAvatar 上传后传入）
	Phone    *string `json:"phone"`    // 手机号明文（前端先调 /user/resolve-phone 用 phoneCode 换好再传）
}

// UpdateMyProfile 编辑当前登录用户本人资料（问题2，只改传入的字段）
func (a *ApiUser) UpdateMyProfile(req UpdateMyProfileReq) (interface{}, error) {
	in := service.MyProfileUpdate{
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Phone:    req.Phone,
	}
	if err := service.GetUserService().UpdateMyProfile(req.ReqUser, in); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
