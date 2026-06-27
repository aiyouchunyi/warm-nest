// Package preset @Author larry
// @Date 2026/06/24
// @Desc 通知路由预设：内置场景→渠道→模板ID 的完整出厂路由（migrate 幂等 seed）。
//
// 为什么模板ID 进预设：服务号模板ID 是平台分配的固定值（非每环境每运营各异的运行时配置），
// 属出厂内置配置——删表重建后应自动恢复、模板消息直接可用，故纳入预设而非留运营手填。
// 运营仍可经管理接口改某场景的渠道/模板ID，migrate 幂等（已存在则跳过）不覆盖改过的值。
//
// NOT_REMIND_ELDER 与 NOT_REMIND_GUARDIAN 共用同一平台模板（漏打卡文案，仅接收方不同）。

package preset

import "warm-nest/internal/model"

// 服务号模板ID（微信公众平台分配，PRD §5.2 模板消息）
const (
	tmplCheckInNotice  = "LEJIwLA7nYgm7BXQiC9R5cuv6LROAOL9BgwTWF1xpec" // 打卡成功通知
	tmplAddressPreheat = "IBPGm9Fvxb7eS7eFP_D-fRWNBJ_xRd59ch7QHryIlsM" // 预热填写地址
	tmplNotRemind      = "2QxRNXkuNOEm4DV0ZVdk3t7qi2rH7lqZNbOFJgt1fWg" // 漏打卡（老人/子女共用）
	tmplBindSuccess    = "vN_fXpU1v4pBbG3WnYDdVTErEKUVMh2h3m9FXOAATnw" // 老人绑定成功
)

// 模板消息点击跳转的小程序页（template 渠道用；运营可在 admin 接口改，migrate 幂等不覆盖）
const (
	pageHome    = "pages/login/index"         // 登录页（多数场景跳这里；小程序入口走登录态分发）
	pageAddress = "pages/child/address/index" // 子女填收货地址页（预热场景）
)

// NotifyRoutes 内置通知路由预设（场景→渠道→模板ID）。
// migrate 按 scene 幂等 seed：新增场景在此追加，已存在的场景不覆盖（运营改过的值保留）。
func NotifyRoutes() []model.NotifyRoute {
	return []model.NotifyRoute{
		{Scene: model.NotifySceneCheckInNotice, Channel: model.NotifyChannelTemplate, TemplateId: tmplCheckInNotice, PagePath: pageHome, Enabled: true},
		{Scene: model.NotifySceneAddressPreheat, Channel: model.NotifyChannelTemplate, TemplateId: tmplAddressPreheat, PagePath: pageAddress, Enabled: true},
		{Scene: model.NotifySceneNotRemindElder, Channel: model.NotifyChannelTemplate, TemplateId: tmplNotRemind, PagePath: pageHome, Enabled: true},
		{Scene: model.NotifySceneNotRemindGuardian, Channel: model.NotifyChannelTemplate, TemplateId: tmplNotRemind, PagePath: pageHome, Enabled: true},
		{Scene: model.NotifySceneBindSuccess, Channel: model.NotifyChannelTemplate, TemplateId: tmplBindSuccess, PagePath: pageHome, Enabled: true},
	}
}
