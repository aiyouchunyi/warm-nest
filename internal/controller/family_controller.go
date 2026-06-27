// Package controller @Author larry
// @Date 2026/06/15
// @Desc 家庭信息 Controller（均需登录）

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// FamilyController 家庭信息 Controller
type FamilyController struct {
	apiFamily *api.ApiFamily
}

var familyController *FamilyController
var familyControllerOnce sync.Once

// GetFamilyController 获取家庭信息 Controller 单例
func GetFamilyController() *FamilyController {
	familyControllerOnce.Do(func() {
		familyController = &FamilyController{apiFamily: api.GetApiFamily()}
	})
	return familyController
}

// Router 注册路由（均需登录）
func (c *FamilyController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1/family")
	group.GET("/view", web.Handler(c.apiFamily.View, widgets.Session))
	group.POST("/profile", web.Handler(c.apiFamily.UpdateProfile, widgets.Session))
	group.GET("/bind-status", web.Handler(c.apiFamily.BindStatus, widgets.Session))
	group.GET("/subscribe-status", web.Handler(c.apiFamily.SubscribeStatus, widgets.Session))
}
