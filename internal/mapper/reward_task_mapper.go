// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 奖励任务规则映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// RewardTaskMapper 奖励任务规则映射器
type RewardTaskMapper struct {
	mysqls.MysqlMapper[model.RewardTask]
}

var rewardTaskMapper *RewardTaskMapper
var rewardTaskOnce sync.Once

// GetRewardTaskMapper 获取奖励任务规则映射器单例
func GetRewardTaskMapper() *RewardTaskMapper {
	rewardTaskOnce.Do(func() {
		rewardTaskMapper = &RewardTaskMapper{
			MysqlMapper: mysqls.NewMapper[model.RewardTask](),
		}
	})
	return rewardTaskMapper
}

// ListEnabled 查全部启用中的奖励任务规则（打卡后达成判定遍历用）
func (m *RewardTaskMapper) ListEnabled() ([]model.RewardTask, error) {
	var list []model.RewardTask
	err := m.Model().Where("enable = ?", true).Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list enabled reward tasks: %w", err)
	}
	return list, nil
}

// GetByTaskKey 按任务编号查规则（种子初始化判重用，找不到返回 nil,nil）
func (m *RewardTaskMapper) GetByTaskKey(taskKey string) (*model.RewardTask, error) {
	var t model.RewardTask
	err := m.Model().Where("task_key = ?", taskKey).First(&t).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get reward task by key %s: %w", taskKey, err)
	}
	return &t, nil
}
