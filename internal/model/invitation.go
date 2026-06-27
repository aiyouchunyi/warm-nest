// Package model @Author larry
// @Date 2026/06/15
// @Desc 守护邀请模型（子女发起 → 老人接受 的两阶段绑定）
//
// 为什么需要邀请表（而非子女直接替老人建账号）：
//   - 绑定关系的前提是「老人有自己的 User」，而 User 必须由老人本人微信登录才能产生（要 openid）。
//     子女发起邀请时老人 User 还不存在，无法直接建 Guardianship（缺 elderUserId）。
//   - 老人的档案（城市/健康/收货地址）应由老人本人补全，子女代填不准也不全。
//   - 故拆两阶段：子女发起邀请（存绑定线索）→ 老人扫码/点链接进小程序、微信登录（产生老人 User）
//     → 接受邀请时才建 ElderProfile + Guardianship，并回填真实 elderUserId。
//
// 关键约束：邀请发起时没有 elderUserId，只能存子女填的「线索」（elderPhone + 预设称呼/提醒时间），
// 接受后才回填 AcceptedElderUserId。InviteCode 是对外传播的口令（小程序码/链接参数），
// 用 rands.NumericN 生成，对外不可枚举。
package model

import "warm-nest/pkg/tool/database"

const (
	InvitationStatusPending   = "PENDING"   // 待接受
	InvitationStatusAccepted  = "ACCEPTED"  // 已接受（已建立守护关系）
	InvitationStatusExpired   = "EXPIRED"   // 已过期
	InvitationStatusCancelled = "CANCELLED" // 已撤销（子女主动取消）
)

// Invitation 守护邀请
type Invitation struct {
	database.Identify
	InvitationId   string `gorm:"size:32;uniqueIndex:u_invitation;comment:业务邀请ID（rands.Numeric 生成）" json:"invitationId"`
	InviteCode     string `gorm:"size:32;uniqueIndex:u_code;comment:邀请口令（小程序码/链接参数，rands.NumericN 生成，对外不可枚举）" json:"inviteCode"`
	GuardianUserId string `gorm:"size:32;index:i_guardian;comment:发起邀请的子女用户ID（已登录，User.UserId）" json:"guardianUserId"`

	// —— 子女预设的绑定线索（接受时落入 Guardianship/ElderProfile）——
	ElderPhone string `gorm:"size:32;comment:子女填的老人手机号（仅线索/预填，接受时落入 ElderProfile.ElderPhone）" json:"elderPhone"`
	Relation   string `gorm:"size:32;comment:子女预设称呼，接受后写入 Guardianship.Relation[复用 Relation* 常量]" json:"relation"`
	RemindTime string `gorm:"size:8;comment:子女预设提醒时间HH:mm（默认09:00），接受后写入 ElderProfile.RemindTime" json:"remindTime"`
	City       string `gorm:"size:64;comment:子女填的老人所在城市（线索，接受时落入 ElderProfile.City 供打卡取天气）" json:"city"`

	Status              string `gorm:"size:32;comment:邀请状态[PENDING:待接受,ACCEPTED:已接受,EXPIRED:已过期,CANCELLED:已撤销]" json:"status"`
	AcceptedElderUserId string `gorm:"size:32;index:i_accepted;comment:老人接受后回填的真实用户ID（接受前为空）" json:"acceptedElderUserId"`
	AcceptedAt          int64  `gorm:"comment:接受时间，毫秒（接受前为0）" json:"acceptedAt"`
	ExpireAt            int64  `gorm:"comment:邀请过期时间，毫秒" json:"expireAt"`

	database.SystemField
}
