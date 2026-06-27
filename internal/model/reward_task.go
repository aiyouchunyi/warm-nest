// Package model @Author larry
// @Date 2026/06/15
// @Desc 奖励任务规则配置模型（定义「满足什么任务 → 奖励什么」）
//
// 设计要点：把奖励抽象成可配置的任务规则，PRD §10 几个待定项（是否需连续达标才发、
// 奖励发放策略）以后都变成改配置而非改表结构：
//   - 条件侧：改 ConditionType/ConditionValue 即可切换达成口径。新增任务类型（邀请/完善资料）
//     加枚举值 + service 判定即可。ConditionValue 暂用单 int，组合条件是未确定的假想场景，不提前上 JSON。
//   - 奖励侧：RewardKind 做类型判别，Quantity 给结构化数量（发货/库存可读），
//     RewardParams(JSON) 放各类型专属属性 —— 未来积分/优惠券/红包等新奖励类型零改表。
//   - 1 期奖励固定为一盒鸡蛋（RewardKindEgg），实物快递发放。
package model

import (
	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/tool/database"
)

// 任务条件类型（达成口径，可扩展：新增任务类型加枚举值 + service 写判定即可，模型不阻塞）
const (
	ConditionMonthlyCheckIn    = "MONTHLY_CHECK_IN"    // 本月打卡满 N 天（对应原型「本月攒蛋进度」）
	ConditionContinuousCheckIn = "CONTINUOUS_CHECK_IN" // 连续打卡 N 天
	ConditionCumulativeCheckIn = "CUMULATIVE_CHECK_IN" // 累计打卡满 N 天（对应「累计 165 天」）
)

// 奖励类型（1 期仅实物鸡蛋；积分/优惠券/红包等类型专属属性放 RewardParams，加类型零改表）
const (
	RewardKindEgg = "EGG" // 一盒鸡蛋（实物，快递发放）
)

// RewardTask 奖励任务规则配置
type RewardTask struct {
	database.Identify
	TaskKey string `gorm:"size:64;uniqueIndex:u_task;comment:任务编号（人工可读配置标识，如 monthly_egg）" json:"taskKey"`
	Name    string `gorm:"size:128;comment:任务名称（如 本月打卡满16天领鸡蛋）" json:"name"`
	Desc    string `gorm:"size:256;comment:任务说明" json:"desc"`

	ConditionType  string `gorm:"size:64;comment:达成条件类型[MONTHLY_CHECK_IN:本月满N天,CONTINUOUS_CHECK_IN:连续N天,CUMULATIVE_CHECK_IN:累计N天]" json:"conditionType"`
	ConditionValue int    `gorm:"comment:达成阈值（天数N）" json:"conditionValue"`

	RewardKind   string           `gorm:"size:64;comment:奖励类型[EGG:一盒鸡蛋]" json:"rewardKind"`
	RewardName   string           `gorm:"size:128;comment:奖励名称（如 安心鸡蛋）" json:"rewardName"`
	RewardSpec   string           `gorm:"size:128;comment:奖励规格描述（给人看，如 30枚/盒）" json:"rewardSpec"`
	Quantity     int              `gorm:"comment:奖励数量（结构化，发货/库存系统读，如 1 盒）" json:"quantity"`
	RewardParams jsons.JSONObject `gorm:"type:json;comment:奖励类型专属属性（EGG:{\"sku\":\"...\"}；积分/优惠券/红包等各自结构）" json:"rewardParams"`

	Enable bool `gorm:"comment:是否启用" json:"enable"`

	database.SystemField
}
