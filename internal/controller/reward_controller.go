// Package controller @Author larry
// @Date 2026/06/15
// @Desc 奖励领取 Controller（均需登录）

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// RewardController 奖励领取 Controller
type RewardController struct {
	apiReward *api.ApiReward
}

var rewardController *RewardController
var rewardControllerOnce sync.Once

// GetRewardController 获取奖励领取 Controller 单例
func GetRewardController() *RewardController {
	rewardControllerOnce.Do(func() {
		rewardController = &RewardController{apiReward: api.GetApiReward()}
	})
	return rewardController
}

// Router 注册路由（均需登录）
func (c *RewardController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1/reward")
	group.GET("/list", web.Handler(c.apiReward.List, widgets.Session))
	group.GET("/detail", web.Handler(c.apiReward.ClaimDetail, widgets.Session))
	group.GET("/address-preheat", web.Handler(c.apiReward.AddressPreheat, widgets.Session)) // 奖励地址预热提醒查询（PRD §6.6.3）
	group.POST("/claim", web.Handler(c.apiReward.Claim, widgets.Session))
}
