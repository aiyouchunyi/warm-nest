// Package controller @Author larry
// @Date 2026/06/15
// @Desc 打卡 Controller（均需登录）

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// CheckInController 打卡 Controller
type CheckInController struct {
	apiCheckIn *api.ApiCheckIn
}

var checkInController *CheckInController
var checkInControllerOnce sync.Once

// GetCheckInController 获取打卡 Controller 单例
func GetCheckInController() *CheckInController {
	checkInControllerOnce.Do(func() {
		checkInController = &CheckInController{apiCheckIn: api.GetApiCheckIn()}
	})
	return checkInController
}

// Router 注册路由（均需登录）
func (c *CheckInController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1/checkin")
	group.POST("/do", web.Handler(c.apiCheckIn.DoCheckIn, widgets.Session))
	group.GET("/today", web.Handler(c.apiCheckIn.TodayStatus, widgets.Session))
	group.GET("/month", web.Handler(c.apiCheckIn.MonthCheckIns, widgets.Session))
	group.GET("/month-overview", web.Handler(c.apiCheckIn.MonthOverview, widgets.Session))
	group.GET("/detail", web.Handler(c.apiCheckIn.CheckInDetail, widgets.Session))
}
