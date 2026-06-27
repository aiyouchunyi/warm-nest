// Package model @Author larry
// @Date 2026/06/15
// @Desc 被守护人档案模型
//
// 设计要点（从 User 拆出）：
//   - User 是纯微信账号；被守护人本人的档案（真名/城市/健康/收货地址/提醒时间）单独成表。
//   - userId 1:1 关联 User，仅当该用户作为被守护人时才有一条 —— 子女账号不背空字段。
//   - 「一个子女守护多个老人」时，是多个老人 User 各自带一份 ElderProfile，互不冗余。
//   - RemindTime 放这里而非 Guardianship：PRD §5.1/§6.5 明确提醒时间=老人的「打卡时间」、
//     双端同源，每个老人只有一个。若放关系边上，一个老人被多子女守护时会出现 N 个冲突的打卡时间。
//   - 与 Guardianship.Relation 的区别：Relation 是「子女对老人的称呼」（爷爷/妈妈，关系视角，每子女可不同），
//     RealName 是老人真实姓名（本人属性，唯一）。
//   - 不存 Age：年龄会随时间失真，且仅展示用；如需展示由 BirthDate 实时算（1 期 PRD 无此需求，暂不加）。
package model

import (
	"warm-nest/pkg/tool/database"
)

// ElderProfile 被守护人档案
//
// 关于「1:1」与「一个老人被多人守护」：1:1 指的是 User ↔ ElderProfile，不是档案 ↔ 守护人数。
// 一个老人被 N 个子女守护，老人仍是同一个 User、同一份档案；多对多体现在 Guardianship 表的 N 条边，
// 档案唯一性不受影响。
//
// 关于收货地址放这里而非 User：奖励（鸡蛋）只发给打卡的老人（RewardClaim.ElderUserId），
// 纯子女账号不产生领取、无需收货，故地址属于「被守护人」而非通用账号。1 期 PRD §7 仅一个收货地址，
// 不做地址簿（YAGNI）；这里存的是「默认收货地址」，下单时带出可改、改后快照进 RewardClaim，
// 历史单不受后续改址影响。
type ElderProfile struct {
	database.Identify
	UserId string `gorm:"size:32;uniqueIndex:u_user;comment:被守护人用户ID（User.UserId，1:1）" json:"userId"`

	RealName   string `gorm:"size:64;comment:真实姓名" json:"realName"`
	City       string `gorm:"size:64;comment:所在城市（用于天气/定位）" json:"city"`
	RemindTime string `gorm:"size:8;comment:打卡/未打卡提醒时间HH:mm（默认09:00，晨起报平安，双端同源）" json:"remindTime"`
	HealthNote string `gorm:"size:512;comment:特殊健康备注（选填，敏感信息）" json:"healthNote"`

	// Address 结构化默认收货地址（json，问题11）：省/市/区/街道/详细 + 收货人 + 收货电话。
	// 收货人/收货电话独立于老人本人，支持子女代收；下单时拍平快照进 RewardClaim。见 address.go。
	Address Address `gorm:"type:json;comment:默认收货地址（结构化json：省市区街道详细+收货人+收货电话）" json:"address"`

	// 联系电话挂档案而非 User：老人/子女的 User.Phone 来自微信登录态、本服务登录不取手机号（恒空），
	// 故联系电话作为「被守护人档案」的可编辑属性维护（双端同源，family 编辑页可改）。
	// ElderPhone   老人本人电话；GuardianPhone 守护人电话——老人有突发情况时按此电话通知家人（PRD §家庭信息）。
	ElderPhone    string `gorm:"size:32;comment:被守护人本人联系电话（档案可编辑，非微信登录手机号）" json:"elderPhone"`
	GuardianPhone string `gorm:"size:32;comment:守护人联系电话（老人突发情况时通知；档案可编辑）" json:"guardianPhone"`

	database.SystemField
}
