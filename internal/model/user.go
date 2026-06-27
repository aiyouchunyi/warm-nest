// Package model @Author larry
// @Date 2026/06/15
// @Desc 用户模型（纯微信账号，不区分角色）
//
// 设计要点：
//   - role（老人/子女）不是人的固有属性，而是「守护关系」的方向，角色由 Guardianship 边表的方向决定。
//     同一个微信用户可在关系 A 中是被守护人、在关系 B 中是守护人（见 PRD §11 多对多预留）。
//   - User 只存账号身份，不存任何角色/档案字段。被守护人本人的档案另见 ElderProfile。
package model

import (
	"warm-nest/pkg/tool/database"
)

const (
	UserStatusNormal   = "NORMAL"   // 正常
	UserStatusDisabled = "DISABLED" // 禁用
)

// 双端视角角色（App 级身份模式；真实身份恒从 Guardianship 边算，这里仅作可用身份与切换偏好的取值）
const (
	RoleElder    = "ELDER"    // 老人端（被守护人视角）
	RoleGuardian = "GUARDIAN" // 子女端（守护人视角）
)

// User 微信用户账号
type User struct {
	database.Identify
	UserId   string `gorm:"size:32;uniqueIndex:u_user;comment:业务用户ID（rands.Numeric 生成，对外不可枚举）" json:"userId"`
	OpenId   string `gorm:"size:64;uniqueIndex:u_open;comment:微信小程序 openid" json:"openId"`
	UnionId  string `gorm:"size:64;index:i_union;comment:微信 unionid（跨小程序/服务号同一用户）" json:"unionId"`
	Phone    string `gorm:"size:32;index:i_phone;comment:登录手机号（明文存储用于绑定匹配，展示时脱敏）" json:"phone"`
	Nickname string `gorm:"size:64;comment:微信昵称" json:"nickname"`
	Avatar   string `gorm:"size:256;comment:头像URL" json:"avatar"`
	Status   string `gorm:"size:32;comment:账号状态[NORMAL:正常,DISABLED:禁用]" json:"status"`

	// LastActiveRole 仅记「上次选的端」作登录默认，可空；非身份本身——身份恒从 Guardianship 边算（见双端视角设计）
	LastActiveRole string `gorm:"size:32;comment:上次使用的端[ELDER:老人端,GUARDIAN:子女端]，仅作切换默认，可空" json:"lastActiveRole"`

	database.SystemField
}
