// Package model @Author larry
// @Date 2026/06/15
// @Desc 守护关系模型（守护人↔被守护人的有向边）
//
// 设计要点（PRD §11 多对多预留）：
//   - 老人与子女业务上是多对多（一个老人可被多个子女守护，一个子女可守护多个老人）。
//   - 独立关系边表维护绑定，不把对方信息冗余进 User 表 —— 这样 B 可同时是 A 的被守护人、C 的守护人。
//   - 方向由 GuardianUserId（守护方=子女视角）+ ElderUserId（被守护方=老人视角）两列决定。
//   - 关系性设定（称呼）挂在边上：不同子女对同一老人可有不同称呼。
//   - 提醒时间 *不* 放这里：它是老人级、双端同源的唯一设置，放在 ElderProfile（见该文件说明）。
//   - 1 期产品仅做一对一，但表结构按多对多设计，无需后续迁移。
package model

import (
	"warm-nest/pkg/tool/database"
)

const (
	GuardianshipStatusActive   = "ACTIVE"   // 守护中
	GuardianshipStatusInactive = "INACTIVE" // 已解除
)

// 关系称呼（子女对被守护人的称呼，对应 PRD §4.3 关系标签）
const (
	RelationMom         = "MOM"          // 妈妈
	RelationDad         = "DAD"          // 爸爸
	RelationGrandpa     = "GRANDPA"      // 爷爷
	RelationGrandma     = "GRANDMA"      // 奶奶
	RelationMaternalGpa = "MATERNAL_GPA" // 外公
	RelationMaternalGma = "MATERNAL_GMA" // 外婆
	RelationOther       = "OTHER"        // 其他
)

// relationText 关系称呼→中文展示（微信模板「用户名/姓名/绑定用户」等字段用，子女视角看老人）
var relationText = map[string]string{
	RelationMom:         "妈妈",
	RelationDad:         "爸爸",
	RelationGrandpa:     "爷爷",
	RelationGrandma:     "奶奶",
	RelationMaternalGpa: "外公",
	RelationMaternalGma: "外婆",
	RelationOther:       "家人",
}

// RelationText 把关系称呼枚举转中文；未知/空值兜底「家人」（不返空，避免模板字段为空被微信拒发）。
func RelationText(relation string) string {
	if t, ok := relationText[relation]; ok {
		return t
	}
	return "家人"
}

// DefaultRemindTime 默认提醒/打卡时间（PRD §4.3：默认 09:00），用于 ElderProfile.RemindTime
const DefaultRemindTime = "09:00"

// Guardianship 守护关系（守护人 -> 被守护人 的有向边）
type Guardianship struct {
	database.Identify
	GuardianshipId string `gorm:"size:32;uniqueIndex:u_relation;comment:业务关系ID（rands.Numeric 生成）" json:"guardianshipId"`
	GuardianUserId string `gorm:"size:32;index:i_guardian;comment:守护人用户ID（子女视角，User.UserId）" json:"guardianUserId"`
	ElderUserId    string `gorm:"size:32;index:i_elder;comment:被守护人用户ID（老人视角，User.UserId）" json:"elderUserId"`

	Relation string `gorm:"size:32;comment:子女对被守护人的称呼[MOM:妈妈,DAD:爸爸,GRANDPA:爷爷,GRANDMA:奶奶,MATERNAL_GPA:外公,MATERNAL_GMA:外婆,OTHER:其他]" json:"relation"`
	Status   string `gorm:"size:32;comment:关系状态[ACTIVE:守护中,INACTIVE:已解除]" json:"status"`

	// ActivatedAt 绑定成立（老人授权接受邀请）的时刻，毫秒（PRD §6.6.1.1 首月 R 起点 / §8.0.2 activated_at）。
	// 与 Invitation.AcceptedAt 同刻写入，但归属关系边——首月奖励窗口/绑定时长按本字段算，不依赖邀请记录。
	ActivatedAt int64 `gorm:"comment:绑定成立时刻（老人授权接受），毫秒——首月奖励窗口起点" json:"activatedAt"`

	database.SystemField
}
