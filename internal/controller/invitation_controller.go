// Package controller @Author larry
// @Date 2026/06/15
// @Desc 守护邀请 Controller（均需登录）

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// InvitationController 守护邀请 Controller
type InvitationController struct {
	apiInvitation *api.ApiInvitation
}

var invitationController *InvitationController
var invitationControllerOnce sync.Once

// GetInvitationController 获取守护邀请 Controller 单例
func GetInvitationController() *InvitationController {
	invitationControllerOnce.Do(func() {
		invitationController = &InvitationController{
			apiInvitation: api.GetApiInvitation(),
		}
	})
	return invitationController
}

// Router 注册路由（均需登录）
func (c *InvitationController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1/invitation")
	group.POST("/create", web.Handler(c.apiInvitation.CreateInvitation, widgets.Session))
	group.POST("/accept", web.Handler(c.apiInvitation.AcceptInvitation, widgets.Session))
	group.POST("/cancel", web.Handler(c.apiInvitation.CancelInvitation, widgets.Session))
	group.GET("/list", web.Handler(c.apiInvitation.ListInvitation, widgets.Session))
	group.GET("/inviter-info", web.Handler(c.apiInvitation.InviterInfo, widgets.Session))
}
