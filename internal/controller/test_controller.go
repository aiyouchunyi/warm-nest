// Package controller @Author larry
// @Date 2026/06/19
// @Desc 测试数据 Controller（供前端自助造数据；临时能力，测完手动删本文件 + controller_center 注册行）
//
// 注意：无鉴权、全环境可用（含 prod）、全 POST（造/改数据）。这是刻意的临时测试通道，
// 上线验证完应及时移除，避免长期暴露造数据接口。路径前缀 /warm-nest/test/v1 与 C 端 /warm-nest/v1 区分。

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"

	"warm-nest/internal/api"
)

// TestController 测试数据 Controller
type TestController struct {
	apiTestData *api.ApiTestData
}

var testController *TestController
var testControllerOnce sync.Once

// GetTestController 获取测试数据 Controller 单例
func GetTestController() *TestController {
	testControllerOnce.Do(func() {
		testController = &TestController{apiTestData: api.GetApiTestData()}
	})
	return testController
}

// Router 注册测试路由（无鉴权，全 POST）
func (c *TestController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/test/v1")
	group.POST("/checkin/seed", web.Handler(c.apiTestData.SeedCheckIns))
	group.POST("/reward/seed-tasks", web.Handler(c.apiTestData.SeedRewardTasks))
	group.POST("/reward/seed-claims", web.Handler(c.apiTestData.SeedRewardClaims))           // 直插指定状态领取记录(测列表多状态/物流展示)
	group.POST("/wipe-all", web.Handler(c.apiTestData.WipeAll))                              // ⚠️全清(含账号)，不可逆，仅测试期
	group.POST("/guardianship/seed", web.Handler(c.apiTestData.SeedGuardianship))            // 造子女→老人守护关系+档案+默认地址
	group.POST("/invitation/seed-pending", web.Handler(c.apiTestData.SeedPendingInvitation)) // 造 PENDING 邀请+子女账号(头像/手机号)
	group.POST("/menu/publish", web.Handler(c.apiTestData.PublishMenu))                      // 发布服务号默认底部菜单(打卡/我的，跳小程序入口页)
	group.POST("/remind/trigger", web.Handler(c.apiTestData.TriggerRemind))                  // 手动触发未打卡提醒轮询(可传 at=HH:mm 模拟时刻)
}
