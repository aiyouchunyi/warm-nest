// Package machine @Author larry
// @Date 2026/06/15
// @Desc 奖励领取状态机服务（manual 动作入口）

package machine

import (
	"sync"

	"warm-nest/pkg/tool/machines/model/dto"
	"warm-nest/pkg/tool/machines/model/manual"
)

// RewardClaimReq 奖励领取状态机操作请求（含登录态、taskId=claimId、action、params）
type RewardClaimReq struct {
	dto.MachineTaskReq
}

// RewardClaimService 奖励领取状态机服务
type RewardClaimService struct {
	rewardClaimMachine *RewardClaimMachine
}

var rewardClaimService *RewardClaimService
var rewardClaimServiceOnce sync.Once

// GetRewardClaimService 获取奖励领取状态机服务单例
func GetRewardClaimService() *RewardClaimService {
	rewardClaimServiceOnce.Do(func() {
		rewardClaimService = &RewardClaimService{
			rewardClaimMachine: GetRewardClaimMachine(),
		}
	})
	return rewardClaimService
}

// Manual 手动推进状态（claimId 定位任务，按 action 流转）
func (s *RewardClaimService) Manual(req RewardClaimReq) error {
	claim, err := s.rewardClaimMachine.FindTask(req.TaskId)
	if err != nil {
		return err
	}
	return s.rewardClaimMachine.Manual(claim, manual.NewEvent(req.MachineTaskReq))
}
