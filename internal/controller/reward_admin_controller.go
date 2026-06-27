// Package controller @Author larry
// @Date 2026/06/15
// @Desc 奖励发放后台 Controller（运营发货/签收，走 machine manual 动作）
//
// 路径前缀 /warm-nest/admin/v1 与 C 端 /warm-nest/v1 区分；挂 widgets.Access（后台权限）。
// machine-task controller 范式：动作化路由 + req.Action 设值 + NoRespHandler。

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"
	"warm-nest/pkg/tool/machines/model/constant"

	"warm-nest/internal/reward/machine"
)

// RewardAdminController 奖励发放后台 Controller
type RewardAdminController struct {
	rewardClaimService *machine.RewardClaimService
}

var rewardAdminController *RewardAdminController
var rewardAdminControllerOnce sync.Once

// GetRewardAdminController 获取奖励发放后台 Controller 单例
func GetRewardAdminController() *RewardAdminController {
	rewardAdminControllerOnce.Do(func() {
		rewardAdminController = &RewardAdminController{
			rewardClaimService: machine.GetRewardClaimService(),
		}
	})
	return rewardAdminController
}

// Router 注册后台路由（均需后台权限 widgets.Access）
func (c *RewardAdminController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/admin/v1/reward-claim")
	// 发货：CLAIMED→SHIPPED，Params 带 expressCompany/expressNo
	group.POST("/ship", web.NoRespHandler(func(req machine.RewardClaimReq) error {
		req.Action = constant.EventActionApprove
		return c.rewardClaimService.Manual(req)
	}, widgets.Access))
	// 签收：SHIPPED→SIGNED
	group.POST("/sign", web.NoRespHandler(func(req machine.RewardClaimReq) error {
		req.Action = constant.EventActionApprove
		return c.rewardClaimService.Manual(req)
	}, widgets.Access))
	// 驳回：当前状态 reject 到失败态
	group.POST("/reject", web.NoRespHandler(func(req machine.RewardClaimReq) error {
		req.Action = constant.EventActionReject
		return c.rewardClaimService.Manual(req)
	}, widgets.Access))
}
