// Package machine @Author larry
// @Date 2026/06/15
// @Desc 签收状态转换器（SHIPPED --approve--> SIGNED，填签收时间）

package transition

import (
	"fmt"
	"sync"
	"time"

	"warm-nest/pkg/tool/machines/impls/transitions"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// RewardClaimSignTransition 签收转换器
type RewardClaimSignTransition struct {
	transitions.DefaultManualTransition
	rewardClaimMapper *mapper.RewardClaimMapper
}

var rewardClaimSignTransition *RewardClaimSignTransition
var rewardClaimSignTransitionOnce sync.Once

// GetRewardClaimSignTransition 获取签收转换器单例
func GetRewardClaimSignTransition() *RewardClaimSignTransition {
	rewardClaimSignTransitionOnce.Do(func() {
		rewardClaimSignTransition = &RewardClaimSignTransition{
			rewardClaimMapper: mapper.GetRewardClaimMapper(),
		}
	})
	return rewardClaimSignTransition
}

// State 当前状态：已发货（待签收）
func (t *RewardClaimSignTransition) State() string {
	return model.ClaimStatusShipped
}

// Config 允许 approve（签收）
func (t *RewardClaimSignTransition) Config() manual.ManualConfig {
	return manual.NewManualConfig(
		manual.WithActions(constant.EventActionApprove),
	)
}

// Execute approve 时填签收时间，转 SIGNED（成功终态）
func (t *RewardClaimSignTransition) Execute(ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	claim := context.ToTask[*model.RewardClaim](ctx)
	claim.SignedAt = time.Now().UnixMilli()
	ctx.Log().Infof("reward claim %s signed", claim.ClaimId)
	return model.ClaimStatusSigned, nil
}

// Succ 持久化签收时间（framework Sync 不回写业务字段）
func (t *RewardClaimSignTransition) Succ(ctx *context.MachineContext, _ manual.ManualEvent, _ string) error {
	claim := context.ToTask[*model.RewardClaim](ctx)
	err := t.rewardClaimMapper.Model().
		Where("claim_id = ?", claim.ClaimId).
		Update("signed_at", claim.SignedAt).Error
	if err != nil {
		return fmt.Errorf("persist sign field %s: %w", claim.ClaimId, err)
	}
	return nil
}
