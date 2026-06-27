// Package model @Author larry
// @Date 2026/06/18
// @Desc 服务号粉丝模型（PRD §8.2 关注服务号引导，关注/取关事件回调维护）
//
// 为什么需要独立粉丝表：
//   - 小程序端无 API 可查「某用户是否关注了服务号」——服务号关注是另一主体（公众平台），
//     小程序 openid 与服务号 openid 不互通，唯一跨主体锚点是 UnionId（同一微信用户）。
//   - 故后端自建粉丝表：监听服务号「关注/取关」事件回调（subscribe/unsubscribe），
//     按服务号 openid + unionId 落库维护关注态。判「老人是否已关注」= 按其 User.UnionId 查本表 Subscribed。
//   - 用途：①未打卡提醒先推老人本人（§8.3，需老人已关注服务号）②绑定成功反馈推子女（§8.0.1.6）
//     ③小程序内关注引导弹窗的「是否已关注」判定（§8.2）。
//
// 关键约束：本表锚点是 UnionId（跨小程序/服务号同一用户），不是小程序 UserId——
// 服务号事件回调来时尚不一定能映射到小程序 User，故 UnionId 必存、UserId 可后补。
package model

import (
	"warm-nest/pkg/tool/database"
)

// Fan 服务号粉丝（关注态）
type Fan struct {
	database.Identify
	UnionId        string `gorm:"size:64;uniqueIndex:u_fan_union;comment:微信 unionid（跨小程序/服务号同一用户锚点，关注判定按此关联 User）" json:"unionId"`
	OfficialOpenId string `gorm:"size:64;index:i_fan_open;comment:服务号侧 openid（与小程序 openid 不同主体，仅服务号内有效）" json:"officialOpenId"`
	Subscribed     bool   `gorm:"column:subscribed;comment:当前是否关注（subscribe事件置true，unsubscribe置false）" json:"subscribed"`
	SubscribeAt    int64  `gorm:"comment:最近一次关注时刻，毫秒" json:"subscribeAt"`
	UnsubscribeAt  int64  `gorm:"comment:最近一次取关时刻，毫秒" json:"unsubscribeAt"`

	database.SystemField
}
