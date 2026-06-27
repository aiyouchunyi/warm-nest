// Package controller @Author larry
// @Date 2026/06/22
// @Desc 通知路由配置后台 Controller（运营改场景→渠道映射，PRD §5.2）
//
// 路径前缀 /warm-nest/admin/v1 与 C 端 /warm-nest/v1 区分；挂 widgets.Access（后台权限），
// 与同目录 reward_admin_controller 一致。
package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// NotifyRouteAdminController 通知路由配置后台 Controller
type NotifyRouteAdminController struct {
	apiNotifyRoute *api.ApiNotifyRoute
}

var notifyRouteAdminController *NotifyRouteAdminController
var notifyRouteAdminControllerOnce sync.Once

// GetNotifyRouteAdminController 获取通知路由配置后台 Controller 单例
func GetNotifyRouteAdminController() *NotifyRouteAdminController {
	notifyRouteAdminControllerOnce.Do(func() {
		notifyRouteAdminController = &NotifyRouteAdminController{
			apiNotifyRoute: api.GetApiNotifyRoute(),
		}
	})
	return notifyRouteAdminController
}

// Router 注册后台路由（均需后台权限 widgets.Access）
func (c *NotifyRouteAdminController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/admin/v1/notify-route")
	group.GET("/list", web.Handler(c.apiNotifyRoute.List, widgets.Access))  // 查全部场景→渠道配置
	group.POST("/save", web.Handler(c.apiNotifyRoute.Save, widgets.Access)) // 改某场景渠道/模板ID/启用态
}
