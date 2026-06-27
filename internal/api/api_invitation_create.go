// Package api @Author larry
// @Date 2026/06/15
// @Desc 发起守护邀请

package api

import (
	"time"

	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/utils/rands"

	"warm-nest/internal/service"
	"warm-nest/internal/storage"
)

// CreateInvitationReq 发起邀请请求（子女端，需登录）
type CreateInvitationReq struct {
	session.Session
	ElderPhone string `json:"elderPhone"` // 老人手机号（线索，接受时落入档案）
	Relation   string `json:"relation"`   // 称呼 MOM/DAD/GRANDPA...
	RemindTime string `json:"remindTime"` // 提醒时间 HH:mm，空则默认 09:00
	City       string `json:"city"`       // 老人所在城市（线索，接受时落入档案供打卡取天气）
	Avatar     string `json:"avatar"`     // 子女头像URL（选填，wx.chooseAvatar 上传后传入，问题5）
	Phone      string `json:"phone"`      // 子女手机号明文（选填，问题5）。前端先调 /user/resolve-phone 用 phoneCode 换得明文再传入，不再直传 code。
}

// CreateInvitation 子女发起守护邀请，返回小程序码
func (a *ApiInvitation) CreateInvitation(req CreateInvitationReq) (interface{}, error) {
	saveWxaCode := func(code []byte) (string, error) {
		return storage.Get().Save(code, "wxacode", rands.Id16()+".png")
	}
	in := service.CreateInvitationInput{
		GuardianUserId: req.ReqUser,
		ElderPhone:     req.ElderPhone,
		Relation:       req.Relation,
		RemindTime:     req.RemindTime,
		City:           req.City,
		Avatar:         req.Avatar,
		Phone:          req.Phone,
	}
	return service.GetInvitationService().Create(in, time.Now().UnixMilli(), saveWxaCode)
}
