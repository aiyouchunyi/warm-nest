// Package machine @Author larry
// @Date 2026/06/15
// @Desc 发货状态转换器（CLAIMED --approve--> SHIPPED，取 Params 填快递快照）

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

// RewardClaimShipTransition 发货转换器
type RewardClaimShipTransition struct {
	transitions.DefaultManualTransition
	rewardClaimMapper *mapper.RewardClaimMapper
}

var rewardClaimShipTransition *RewardClaimShipTransition
var rewardClaimShipTransitionOnce sync.Once

// GetRewardClaimShipTransition 获取发货转换器单例
func GetRewardClaimShipTransition() *RewardClaimShipTransition {
	rewardClaimShipTransitionOnce.Do(func() {
		rewardClaimShipTransition = &RewardClaimShipTransition{
			rewardClaimMapper: mapper.GetRewardClaimMapper(),
		}
	})
	return rewardClaimShipTransition
}

// State 当前状态：已领取（待发货）
func (t *RewardClaimShipTransition) State() string {
	return model.ClaimStatusClaimed
}

// Config 允许 approve（发货）/ reject
func (t *RewardClaimShipTransition) Config() manual.ManualConfig {
	return manual.NewManualConfig(
		manual.WithActions(constant.EventActionApprove, constant.EventActionReject),
	)
}

// Execute approve 时取 Params 的快递公司/单号填快照，转 SHIPPED
func (t *RewardClaimShipTransition) Execute(ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	if event.Action != constant.EventActionApprove {
		return constant.MachineFailed, nil
	}
	claim := context.ToTask[*model.RewardClaim](ctx)
	company, _ := event.Params["expressCompany"].(string)
	no, _ := event.Params["expressNo"].(string)
	if company == "" || no == "" {
		return "", fmt.Errorf("ship reward claim %s: expressCompany/expressNo required in params", claim.ClaimId)
	}
	claim.ExpressCompany = company
	claim.ExpressNo = no
	claim.ShippedAt = time.Now().UnixMilli()
	ctx.Log().Infof("reward claim %s shipped, express=%s/%s", claim.ClaimId, company, no)
	return model.ClaimStatusShipped, nil
}

// Succ 持久化业务字段（framework 的 Sync 只回写状态机字段，业务字段须在此显式落库）
func (t *RewardClaimShipTransition) Succ(ctx *context.MachineContext, _ manual.ManualEvent, _ string) error {
	claim := context.ToTask[*model.RewardClaim](ctx)
	err := t.rewardClaimMapper.Model().
		Where("claim_id = ?", claim.ClaimId).
		Updates(map[string]interface{}{
			"express_company": claim.ExpressCompany,
			"express_no":      claim.ExpressNo,
			"shipped_at":      claim.ShippedAt,
		}).Error
	if err != nil {
		return fmt.Errorf("persist ship fields %s: %w", claim.ClaimId, err)
	}
	return nil
}
