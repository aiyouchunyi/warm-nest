// Package api @Author larry
// @Date 2026/06/15
// @Desc 家庭信息接口（双端同源）

package api

import (
	"sync"

	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// ApiFamily 家庭信息接口
type ApiFamily struct{}

var apiFamily *ApiFamily
var apiFamilyOnce sync.Once

// GetApiFamily 获取家庭信息接口单例
func GetApiFamily() *ApiFamily {
	apiFamilyOnce.Do(func() {
		apiFamily = &ApiFamily{}
	})
	return apiFamily
}

// ViewFamilyReq 家庭信息请求（需登录）
type ViewFamilyReq struct {
	session.Session
}

// View 双端视图（viewerRole 由后端按边判定，文案前端渲染）
func (a *ApiFamily) View(req ViewFamilyReq) (interface{}, error) {
	return service.GetFamilyService().ViewFamily(req.ReqUser)
}

// UpdateProfileReq 编辑老人档案请求（需登录，整页全量覆盖）
//
// 双端同源：老人本人或其 active 守护人均可调用（权限在 service 校验）。
// 全量覆盖——前端须先从 view 回填当前值再提交，未传字段会被覆盖为空。
// relation（问题1）挂守护边，仅子女端调用时生效。
// ⚠️ 收货地址已迁到独立地址簿（问题3），不在本接口维护——增删改/设默认走 /warm-nest/v1/address/*。
type UpdateProfileReq struct {
	session.Session
	ElderUserId   string `json:"elderUserId" validate:"required"`
	RealName      string `json:"realName"`
	City          string `json:"city"`
	RemindTime    string `json:"remindTime" validate:"required"` // HH:mm，老人级唯一设置
	HealthNote    string `json:"healthNote"`                     // 敏感信息，勿打日志
	Relation      string `json:"relation"`                       // 子女对老人称呼（挂守护边，老人端调用忽略）
	ElderPhone    string `json:"elderPhone"`                     // 老人本人联系电话（可改）
	GuardianPhone string `json:"guardianPhone"`                  // 守护人联系电话（老人突发情况时通知，可改）
}

// UpdateProfile 编辑老人档案（含提醒时间，双端同源生效）
func (a *ApiFamily) UpdateProfile(req UpdateProfileReq) (interface{}, error) {
	in := service.ProfileUpdate{
		RealName:      req.RealName,
		City:          req.City,
		RemindTime:    req.RemindTime,
		HealthNote:    req.HealthNote,
		Relation:      req.Relation,
		ElderPhone:    req.ElderPhone,
		GuardianPhone: req.GuardianPhone,
	}
	if err := service.GetFamilyService().UpdateProfile(req.ReqUser, req.ElderUserId, in); err != nil {
		return nil, err
	}
	return map[string]bool{"ok": true}, nil
}
