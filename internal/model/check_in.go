// Package model @Author larry
// @Date 2026/06/15
// @Desc 打卡记录模型（PRD §3.1 老人每日拍照打卡）
//
// 设计要点：
//   - 打卡数据永远绑定「被守护人自己」（ElderUserId），与守护关系无关 —— 子女通过关系边查到对方打卡流。
//   - 每自然日仅一条有效打卡：CheckInDate（YYYY-MM-DD）+ ElderUserId 唯一索引兜底防重。
//   - 打卡状态以服务端时间的自然日为周期（PRD §3.1.1 交互规则2、§5.3 以服务端时间为准）。
package model

import (
	"warm-nest/pkg/tool/database"
)

const (
	CheckInKindNormal = "NORMAL" // 当日正常打卡（CheckInDate == 创建当天）
	CheckInKindRepair = "REPAIR" // 补打卡（补过去某一天，CheckInDate < 创建当天）
)

// CheckIn 打卡记录
//
// 关于补打卡（PRD §10 待定项，模型先把能力建到位）：
//   - 正常打卡：CheckInDate（打哪天）== CreatedAt 所在自然日。
//   - 补打卡：老人漏打卡后补一张过去某天的卡，此时 CheckInDate（补的是哪天，过去）
//     与 CreatedAt（实际补卡的时刻，今天）必然不同 —— 这正是要独立建模 CheckInDate 的原因，
//     不能用 CreatedAt 推算打卡归属日。
//   - Kind 区分正常/补卡；是否计入连续天数、是否触发奖励，由业务层按补卡策略决定（不写死在模型）。
type CheckIn struct {
	database.Identify
	CheckInId   string `gorm:"size:32;uniqueIndex:u_checkin;comment:业务打卡ID（rands.Numeric 生成）" json:"checkInId"`
	ElderUserId string `gorm:"size:32;uniqueIndex:u_user_date,priority:1;index:i_elder;comment:被守护人用户ID（打卡归属人）" json:"elderUserId"`
	CheckInDate string `gorm:"size:16;uniqueIndex:u_user_date,priority:2;comment:打卡归属自然日YYYY-MM-DD（补卡时为过去某天，防一日多卡）" json:"checkInDate"`
	Kind        string `gorm:"size:32;comment:打卡类型[NORMAL:当日正常,REPAIR:补打卡]" json:"kind"`

	PhotoUrl string `gorm:"size:512;comment:打卡照片URL（敏感数据，加密存储）" json:"photoUrl"`
	Weather  string `gorm:"size:64;comment:打卡时天气文案（如 小雨 22°C）" json:"weather"`
	City     string `gorm:"size:64;comment:打卡所在城市" json:"city"`

	database.SystemField
}
