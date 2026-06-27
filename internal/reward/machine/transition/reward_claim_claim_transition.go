// Package machine @Author larry
// @Date 2026/06/15
// @Desc 领取状态转换器（PENDING --approve--> CLAIMED，填领取时间）

package transition

import (
	"fmt"
	"sync"
	"time"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/machines/impls/transitions"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"

	wnconst "warm-nest/internal/constant"
	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
	"warm-nest/internal/service"
)

// RewardClaimClaimTransition 领取转换器
type RewardClaimClaimTransition struct {
	transitions.DefaultManualTransition
	rewardClaimMapper *mapper.RewardClaimMapper
	addressService    *service.AddressService
}

var rewardClaimClaimTransition *RewardClaimClaimTransition
var rewardClaimClaimTransitionOnce sync.Once

// GetRewardClaimClaimTransition 获取领取转换器单例
func GetRewardClaimClaimTransition() *RewardClaimClaimTransition {
	rewardClaimClaimTransitionOnce.Do(func() {
		rewardClaimClaimTransition = &RewardClaimClaimTransition{
			rewardClaimMapper: mapper.GetRewardClaimMapper(),
			addressService:    service.GetAddressService(),
		}
	})
	return rewardClaimClaimTransition
}

// State 当前状态：待领取
func (t *RewardClaimClaimTransition) State() string {
	return model.ClaimStatusPending
}

// Config 允许 approve（领取）/ reject
func (t *RewardClaimClaimTransition) Config() manual.ManualConfig {
	return manual.NewManualConfig(
		manual.WithActions(constant.EventActionApprove, constant.EventActionReject),
	)
}

// Execute approve 时按 addressId 选地址→校验收货信息→填领取时间+收货快照，转 CLAIMED；reject 转失败态。
// 收货校验（问题9 细化）：实物奖励领取必须有完整收货信息（地址主体/收货人/收货电话），
// 缺哪项返对应结构化错误码，前端据此跳「编辑收货地址」页补对应字段。
// 收货快照来源（问题3 连锁，地址簿）：从地址簿按 Params["addressId"] 取——传了用该条（校验属本老人），
// 未传取该老人默认地址；整条 model.Address 快照进 ReceiverAddress，后续改址不影响本历史单。
func (t *RewardClaimClaimTransition) Execute(ctx *context.MachineContext, event manual.ManualEvent) (string, error) {
	if event.Action != constant.EventActionApprove {
		return constant.MachineFailed, nil
	}
	claim := context.ToTask[*model.RewardClaim](ctx)

	// addressId 可空（同 ship transition 取 expressNo 范式）：未传则取默认地址
	addressId, _ := event.Params["addressId"].(string)
	addr, err := t.addressService.ResolveShippingAddress(claim.ElderUserId, addressId)
	if err != nil {
		ctx.Log().WithField("claimId", claim.ClaimId).Warn("claim rejected: resolve shipping address failed")
		return "", err
	}
	// 收货信息完整性校验：按缺失字段返对应错误码（前端跳编辑页补对应项）
	switch addr.MissingShippingField() {
	case "address":
		ctx.Log().WithField("claimId", claim.ClaimId).Warn("claim rejected: receiver address required")
		return "", errors.NewWithArgs(wnconst.ErrAddressRequired)
	case "receiverName":
		ctx.Log().WithField("claimId", claim.ClaimId).Warn("claim rejected: receiver name required")
		return "", errors.NewWithArgs(wnconst.ErrReceiverRequired)
	case "receiverPhone":
		ctx.Log().WithField("claimId", claim.ClaimId).Warn("claim rejected: receiver phone required")
		return "", errors.NewWithArgs(wnconst.ErrReceiverPhoneRequired)
	}
	claim.ReceiverAddress = addr // 整体快照（结构化），后续改址不影响本历史单
	claim.ClaimedAt = time.Now().UnixMilli()
	ctx.Log().Infof("reward claim %s claimed", claim.ClaimId)
	return model.ClaimStatusClaimed, nil
}

// Succ 持久化领取时间 + 收货信息快照
func (t *RewardClaimClaimTransition) Succ(ctx *context.MachineContext, _ manual.ManualEvent, _ string) error {
	claim := context.ToTask[*model.RewardClaim](ctx)
	err := t.rewardClaimMapper.Model().
		Where("claim_id = ?", claim.ClaimId).
		Updates(map[string]interface{}{
			"claimed_at":       claim.ClaimedAt,
			"receiver_address": claim.ReceiverAddress,
		}).Error
	if err != nil {
		return fmt.Errorf("persist claim field %s: %w", claim.ClaimId, err)
	}
	return nil
}
