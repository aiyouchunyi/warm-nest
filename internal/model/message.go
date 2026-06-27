// Package model @Author larry
// @Date 2026/06/15
// @Desc 小程序内消息流模型（PRD §4.1 子女端打卡消息列表 / §5 提醒）
//
// 设计要点：
//   - 1 期仅落小程序内消息流；微信服务号模板消息下发状态表暂不建
//     （PRD §10 推送方案本身待定，待方案确定后再补 PushRecord 表）。
//   - 接收方是子女（ReceiverUserId），消息中的老人是 ElderUserId，
//     关联触发它的打卡（RefCheckInId，未打卡提醒时为空）。
//   - 标题/正文不落库整句：消息都是模板化的，存结构化参数 Params（老人称呼/天气/打卡时间等），
//     前端按 Type + Params 渲染。这也对齐微信服务号模板消息的「模板ID + 参数」下发模式，
//     未来补 PushRecord 推送可直接复用同一份 Params，避免存死整句文案。
//   - 幂等防重：MessageId 是随机串挡不住重复写入（打卡接口重试 / Kafka 重投 / 提醒任务重跑都会重发）。
//     用 DedupKey（业务语义键）唯一索引兜底——打卡通知 checkin:{checkInId}:{receiverUserId}，
//     未打卡提醒 remind:{elderUserId}:{receiverUserId}:{date}，写入冲突即幂等跳过。
package model

import (
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/database"
)

// 消息类型（PRD §5.2 / §8.0）
const (
	MessageTypeCheckIn        = "CHECK_IN"        // 打卡通知（老人完成打卡）
	MessageTypeNotRemind      = "NOT_REMIND"      // 未打卡提醒（到约定时间仍未打卡）
	MessageTypeBindSuccess    = "BIND_SUCCESS"    // 绑定成功反馈（老人授权后通知子女「✅ 已成功绑定 XX」，PRD §8.0.1.6）
	MessageTypeAddressPreheat = "ADDRESS_PREHEAT" // 奖励地址预热提醒（每月25号、收货地址空→提醒子女补地址，PRD §6.6.3）
)

const (
	MessageUnread = false // 未读
	MessageRead   = true  // 已读
)

// Message 小程序内消息
type Message struct {
	database.Identify
	MessageId      string `gorm:"size:32;uniqueIndex:u_message;comment:业务消息ID（rands.Numeric 生成）" json:"messageId"`
	DedupKey       string `gorm:"size:128;uniqueIndex:u_dedup;comment:幂等去重键（按类型生成，写入冲突即跳过，防重试/重投/重跑重复发）" json:"dedupKey"`
	ReceiverUserId string `gorm:"size:32;index:i_receiver;comment:接收方用户ID（子女）" json:"receiverUserId"`
	ElderUserId    string `gorm:"size:32;index:i_elder;index:i_elder_bizdate,priority:1;comment:被守护人用户ID（消息中的老人）" json:"elderUserId"`

	Type         string           `gorm:"size:32;comment:消息类型[CHECK_IN:打卡通知,NOT_REMIND:未打卡提醒,BIND_SUCCESS:绑定成功,ADDRESS_PREHEAT:地址预热提醒]" json:"type"`
	Params       jsons.JSONObject `gorm:"type:json;comment:模板渲染参数（如{\"elderName\":\"爷爷\",\"weather\":\"小雨22°C\",\"checkInAt\":\"08:50\"}），前端按Type渲染标题正文" json:"params"`
	RefCheckInId string           `gorm:"size:32;comment:关联打卡ID（打卡通知指向该次打卡，未打卡提醒为空）" json:"refCheckInId"`

	// BizDate 消息业务归属日（YYYY-MM-DD，与 CheckIn.CheckInDate 同口径，Asia/Shanghai）。
	// 未打卡提醒(NOT_REMIND)落本字段，供老人补打卡时按 (elderUserId, bizDate) 精准定位当日已发提醒并标「已打卡」(§5.3)。
	BizDate string `gorm:"size:10;index:i_elder_bizdate,priority:2;comment:消息业务归属日YYYY-MM-DD（未打卡提醒用，补打卡标记按此定位）" json:"bizDate"`
	// ElderCheckedAt 老人当日补打卡后对本条未打卡提醒的标记时刻（毫秒，0=未标记）。
	// §5.3：提醒已发出不撤回，仅标「已打卡」让消息流不再显示为有效待处理。前端据 >0 渲染「老人已打卡」。
	ElderCheckedAt int64 `gorm:"comment:老人当日补打卡标记时刻，毫秒（0=未标记，仅未打卡提醒会被标记）" json:"elderCheckedAt"`

	IsRead bool  `gorm:"column:is_read;comment:是否已读" json:"isRead"` // 列名避开 MySQL 保留字 read
	ReadAt int64 `gorm:"comment:已读时间，毫秒" json:"readAt"`

	database.SystemField
}
