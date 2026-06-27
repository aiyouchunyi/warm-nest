// Package migrate @Author larry
// @Date 2026/06/24
// @Desc 奖励任务规则初始化（预设数据，PRD §6 奖励规则出厂内置）。
//
// 为什么需要：reward_task 是预设数据（出厂内置奖励规则），原仅靠测试接口手动 seed，且被 WipeAll 清，
// 清库后规则丢失需手工重灌。改为 migrate 幂等 seed（不存在则插入），与 notify_route 同范式。
//
// 幂等：按 TaskKey 判重，已存在的规则跳过（不覆盖运营/测试改过的值）。

package migrate

import (
	"fmt"
	"sync"

	"warm-nest/internal/mapper"
	"warm-nest/internal/preset"
)

// RewardTaskMigrate 奖励任务规则初始化
type RewardTaskMigrate struct {
	rewardTaskMapper *mapper.RewardTaskMapper
}

var rewardTaskMigrate *RewardTaskMigrate
var rewardTaskMigrateOnce sync.Once

// GetRewardTaskMigrate 获取奖励任务规则初始化单例
func GetRewardTaskMigrate() *RewardTaskMigrate {
	rewardTaskMigrateOnce.Do(func() {
		rewardTaskMigrate = &RewardTaskMigrate{
			rewardTaskMapper: mapper.GetRewardTaskMapper(),
		}
	})
	return rewardTaskMigrate
}

// Migrate 幂等 seed 预设奖励规则（按 TaskKey 判重，已存在跳过）
func (m *RewardTaskMigrate) Migrate() error {
	for _, task := range preset.RewardTasks() {
		existing, err := m.rewardTaskMapper.GetByTaskKey(task.TaskKey)
		if err != nil {
			return fmt.Errorf("reward task migrate check %s: %w", task.TaskKey, err)
		}
		if existing != nil {
			continue // 幂等：已有规则不覆盖
		}
		t := task
		if err := m.rewardTaskMapper.Create(&t); err != nil {
			return fmt.Errorf("reward task migrate seed %s: %w", task.TaskKey, err)
		}
	}
	return nil
}
