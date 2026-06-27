// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 领取记录映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// RewardClaimMapper 领取记录映射器
type RewardClaimMapper struct {
	mysqls.MysqlMapper[model.RewardClaim]
}

var rewardClaimMapper *RewardClaimMapper
var rewardClaimOnce sync.Once

// GetRewardClaimMapper 获取领取记录映射器单例
func GetRewardClaimMapper() *RewardClaimMapper {
	rewardClaimOnce.Do(func() {
		rewardClaimMapper = &RewardClaimMapper{
			MysqlMapper: mysqls.NewMapper[model.RewardClaim](),
		}
	})
	return rewardClaimMapper
}

// ExistByUserTaskPeriod 判某人某任务某周期是否已发放（防重发，配合唯一索引）
func (m *RewardClaimMapper) ExistByUserTaskPeriod(elderUserId, taskKey, periodKey string) (bool, error) {
	var cnt int64
	err := m.Model().
		Where("elder_user_id = ? AND task_key = ? AND period_key = ?", elderUserId, taskKey, periodKey).
		Count(&cnt).Error
	if err != nil {
		return false, fmt.Errorf("exist reward claim %s/%s/%s: %w", elderUserId, taskKey, periodKey, err)
	}
	return cnt > 0, nil
}

// ListByUser 查某人全部领取记录（按创建倒序）
func (m *RewardClaimMapper) ListByUser(elderUserId string) ([]model.RewardClaim, error) {
	var list []model.RewardClaim
	err := m.Model().
		Where("elder_user_id = ?", elderUserId).
		Order("created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("list reward claim by %s: %w", elderUserId, err)
	}
	return list, nil
}

// GetByClaimId 按业务领取ID查记录（找不到返回 nil,nil）
func (m *RewardClaimMapper) GetByClaimId(claimId string) (*model.RewardClaim, error) {
	var c model.RewardClaim
	err := m.Model().Where("claim_id = ?", claimId).First(&c).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get reward claim by id %s: %w", claimId, err)
	}
	return &c, nil
}
