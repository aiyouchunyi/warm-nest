// Package controller @Author larry
// @Date 2026/06/21
// @Desc 收货地址簿 Controller（问题3，均需登录）

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// AddressController 收货地址簿 Controller
type AddressController struct {
	apiAddress *api.ApiAddress
}

var addressController *AddressController
var addressControllerOnce sync.Once

// GetAddressController 获取收货地址簿 Controller 单例
func GetAddressController() *AddressController {
	addressControllerOnce.Do(func() {
		addressController = &AddressController{apiAddress: api.GetApiAddress()}
	})
	return addressController
}

// Router 注册路由（均需登录）：查询用 GET，增改删/设默认用 POST
func (c *AddressController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1/address")
	group.GET("/list", web.Handler(c.apiAddress.List, widgets.Session))
	group.POST("/add", web.Handler(c.apiAddress.Add, widgets.Session))
	group.POST("/update", web.Handler(c.apiAddress.Update, widgets.Session))
	group.POST("/delete", web.Handler(c.apiAddress.Delete, widgets.Session))
	group.POST("/set-default", web.Handler(c.apiAddress.SetDefault, widgets.Session))
}
