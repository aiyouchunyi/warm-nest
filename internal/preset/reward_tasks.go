// Package preset @Author larry
// @Date 2026/06/24
// @Desc 预设数据（启动时由 migrate 幂等 seed，不存在则插入）的唯一真相源。
//
// 约定（与运行时业务数据区分）：
//   - 预设数据 = 系统出厂内置、业务运行不可或缺的基础配置（奖励规则、通知场景路由…）。
//   - 落库方式：migrate 幂等调用本包的 RewardTasks()/NotifyScenes()，「按 key 查、不存在则插入」，
//     重跑/重启不覆盖运营改过的值。
//   - 清理边界：测试清理接口（WipeAll）**不得清理预设数据表**（reward_task/notify_route），
//     避免清库后业务规则丢失需手工重灌。运行时业务数据（user/checkin/message…）才可被 WipeAll 清。
package preset

import (
	"warm-nest/pkg/kinds/jsons"

	"warm-nest/internal/model"
)

// RewardTasks 预设奖励任务规则（出厂内置，migrate 幂等 seed）。
// 按 TaskKey 判重插入；新增规则在此追加一条即可，migrate 会自动补齐缺失项。
func RewardTasks() []model.RewardTask {
	return []model.RewardTask{
		{
			TaskKey:        "monthly_egg",
			Name:           "本月打卡领鸡蛋",
			Desc:           "本月打卡满（当月天数-3）天，领一盒安心鸡蛋",
			ConditionType:  model.ConditionMonthlyCheckIn,
			ConditionValue: 0, // monthly 达标线动态按月算，不读此值
			RewardKind:     model.RewardKindEgg,
			RewardName:     "安心鸡蛋",
			RewardSpec:     "30枚/盒",
			Quantity:       1,
			RewardParams:   jsons.JSONObject{"sku": "EGG-30"},
			Enable:         true,
		},
		{
			TaskKey:        "continuous_egg_7",
			Name:           "连续打卡7天领鸡蛋",
			Desc:           "连续打卡满 7 天，领一盒安心鸡蛋（测试便利低门槛任务）",
			ConditionType:  model.ConditionContinuousCheckIn,
			ConditionValue: 7,
			RewardKind:     model.RewardKindEgg,
			RewardName:     "安心鸡蛋",
			RewardSpec:     "30枚/盒",
			Quantity:       1,
			RewardParams:   jsons.JSONObject{"sku": "EGG-30"},
			Enable:         true,
		},
	}
}
