// Package model @Author larry
// @Date 2026/06/15
// @Desc 领取记录 / 任务达成流水模型（PRD §6.1/§6.2 领取记录与详情）
//
// 设计要点：
//   - 用户每达成一次 RewardTask 产生一条记录，归属被守护人（ElderUserId）。
//   - PeriodKey 标识达成周期（如 2026-06 / cumulative-30），配合
//     ElderUserId+TaskKey+PeriodKey 唯一索引防重复发放。
//   - reward* 三字段是 *快照*（非关联 RewardTask 实时读）：领取记录是订单流水，
//     规则以后会改（30枚→50枚、改名），历史单必须保留达成当时实际发放的内容。
//   - 收货信息同样下单时从 ElderProfile 拷一份快照，避免老人事后改地址影响历史单。
//   - 状态流转（待领→已领→已发货→已签收）走 machine 状态机引擎：
//     嵌 task.MachineTask 提供 Status + 流转历史；C 端领取、后台发货/签收均通过 machine 的 manual 动作推进，
//     由 reward/machine 包定义 transitions。各状态业务时间戳（ClaimedAt 等）由对应 transition 落。
package model

import (
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/database"
	"warm-nest/pkg/tool/machines/model/task"
)

// 领取状态（PRD §6.1：待领取 / 已领取 / 已发货 / 已签收），作为 MachineTask.Status 取值
const (
	ClaimStatusPending = "PENDING" // 待领取（达成后初始态）
	ClaimStatusClaimed = "CLAIMED" // 已领取（老人/子女确认领取，待发货）
	ClaimStatusShipped = "SHIPPED" // 已发货（已填快递单号）
	ClaimStatusSigned  = "SIGNED"  // 已签收（终态）
)

// RewardClaim 领取记录 / 达成流水
type RewardClaim struct {
	database.Identify
	ClaimId     string `gorm:"size:32;uniqueIndex:u_claim;comment:业务领取ID（rands.Numeric 生成）" json:"claimId"`
	ElderUserId string `gorm:"size:32;uniqueIndex:u_user_task_period,priority:1;index:i_elder;comment:被守护人用户ID（领取归属人）" json:"elderUserId"`
	TaskKey     string `gorm:"size:64;uniqueIndex:u_user_task_period,priority:2;comment:命中的奖励任务编号（RewardTask.TaskKey）" json:"taskKey"`
	PeriodKey   string `gorm:"size:64;uniqueIndex:u_user_task_period,priority:3;comment:达成周期标识（如 2026-06 / cumulative-30），防重复发放" json:"periodKey"`

	// —— 奖励内容快照（达成时从 RewardTask 拷贝，规则变更不影响历史单）——
	RewardKind   string           `gorm:"size:64;comment:奖励类型快照[EGG:一盒鸡蛋]" json:"rewardKind"`
	RewardName   string           `gorm:"size:128;comment:奖励名称快照" json:"rewardName"`
	RewardSpec   string           `gorm:"size:128;comment:奖励规格快照（如 30枚/盒）" json:"rewardSpec"`
	Quantity     int              `gorm:"comment:奖励数量快照（发货/库存读）" json:"quantity"`
	RewardParams jsons.JSONObject `gorm:"type:json;comment:奖励类型专属属性快照（与 RewardTask.RewardParams 同结构）" json:"rewardParams"`

	AchievedSnap jsons.JSONObject `gorm:"type:json;comment:达成快照（达成时进度，如连续天数/月份）" json:"achievedSnap"`

	// —— 物流（PRD §6.2，待领取时为空）——
	ExpressCompany string           `gorm:"size:64;comment:快递公司" json:"expressCompany"`
	ExpressNo      string           `gorm:"size:64;index:i_express;comment:快递单号" json:"expressNo"`
	ExpressDetail  jsons.JSONObject `gorm:"type:json;comment:物流轨迹缓存（开放平台查询结果，避免重复查；签收后不再刷新）" json:"expressDetail"`
	ExpressSyncAt  int64            `gorm:"comment:物流轨迹最后同步时刻，毫秒" json:"expressSyncAt"`

	// —— 收货信息快照（下单时从 ElderProfile.Address 整体拷贝，结构化 json，与档案地址同构）——
	// 与 ElderProfile.Address 同为 model.Address 结构：下单时整体快照、后续改址不影响历史单；
	// 用 json 而非拍平字符串，取数统一（直接整拷）、保留省市区结构供发货/统计（与档案地址一致）。
	ReceiverAddress Address `gorm:"type:json;comment:收货信息快照（结构化json：省市区街道详细+收货人+收货电话）" json:"receiverAddress"`

	// —— 各状态时间节点（PRD §6.2；达成时间即 SystemField.CreatedAt）——
	ClaimedAt int64 `gorm:"comment:领取时间，毫秒" json:"claimedAt"`
	ShippedAt int64 `gorm:"comment:发货时间，毫秒" json:"shippedAt"`
	SignedAt  int64 `gorm:"comment:签收时间，毫秒" json:"signedAt"`

	task.MachineTask // 提供 Status + 流转历史 StatusInfos（machine 引擎驱动）
	database.SystemField
}

// GetIdentify 状态机任务标识（接入 machine 状态机引擎，按 ClaimId 定位任务）
func (c *RewardClaim) GetIdentify() task.TaskIdentify {
	return task.TaskIdentify{
		TaskName:  "奖励领取发放",
		TaskField: "claimId",
		TaskId:    c.ClaimId,
	}
}
