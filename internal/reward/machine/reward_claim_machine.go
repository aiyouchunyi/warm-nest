// Package machine @Author larry
// @Date 2026/06/15
// @Desc 奖励领取发放状态机（PENDING→CLAIMED→SHIPPED→SIGNED，人工驱动）
//
// 流转：
//   - PENDING  --approve--> CLAIMED（C 端用户领取）
//   - CLAIMED  --approve--> SHIPPED（后台发货，自定义 transition 取 Params 填快递）
//   - SHIPPED  --approve--> SIGNED （后台/物流签收，自定义 transition 填签收时间）
// 各状态可 reject 到 MachineFailed。SIGNED 为成功终态。

package machine

import (
	"sync"

	"warm-nest/pkg/tool/machines"
	"warm-nest/pkg/tool/machines/impls"

	"warm-nest/internal/model"
	"warm-nest/internal/reward/machine/transition"
)

// RewardClaimMachine 奖励领取发放状态机
type RewardClaimMachine struct {
	*machines.Machine[*model.RewardClaim]
}

var rewardClaimMachine *RewardClaimMachine
var rewardClaimMachineOnce sync.Once

// GetRewardClaimMachine 获取奖励领取状态机单例
func GetRewardClaimMachine() *RewardClaimMachine {
	rewardClaimMachineOnce.Do(func() {
		rewardClaimMachine = &RewardClaimMachine{
			machines.NewMachine[*model.RewardClaim](
				"RewardClaimMachine",
				impls.NewDefaultTransition(
					// PENDING：用户领取（自定义 transition 填领取时间）
					transition.GetRewardClaimClaimTransition(),
					// CLAIMED：后台发货（自定义 transition 填快递快照）
					transition.GetRewardClaimShipTransition(),
					// SHIPPED：签收（自定义 transition 填签收时间）
					transition.GetRewardClaimSignTransition(),
				),
			),
		}
	})
	return rewardClaimMachine
}
