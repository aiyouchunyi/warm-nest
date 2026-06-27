// Package controller @Author larry
// @Date 2026/06/15
// @Desc 用户 Controller（登录免鉴权，其余需登录）

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// UserController 用户 Controller
type UserController struct {
	apiUser *api.ApiUser
}

var userController *UserController
var userControllerOnce sync.Once

// GetUserController 获取用户 Controller 单例
func GetUserController() *UserController {
	userControllerOnce.Do(func() {
		userController = &UserController{
			apiUser: api.GetApiUser(),
		}
	})
	return userController
}

// Router 注册路由
func (c *UserController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1/user")
	// 登录：免鉴权（不挂 widget）
	group.POST("/login", web.Handler(c.apiUser.Login))
	// 以下均需登录
	group.POST("/resolve-phone", web.Handler(c.apiUser.ResolvePhone, widgets.Session)) // 凭 phoneCode 换手机号明文（问题5）
	group.GET("/profile", web.Handler(c.apiUser.MyProfile, widgets.Session))           // 查本人资料（问题2）
	group.POST("/profile", web.Handler(c.apiUser.UpdateMyProfile, widgets.Session))    // 编辑本人资料（问题2）
}
