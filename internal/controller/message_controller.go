// Package controller @Author larry
// @Date 2026/06/15
// @Desc 消息 Controller（均需登录）

package controller

import (
	"sync"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web"
	"warm-nest/pkg/app/web/widgets"

	"warm-nest/internal/api"
)

// MessageController 消息 Controller
type MessageController struct {
	apiMessage *api.ApiMessage
}

var messageController *MessageController
var messageControllerOnce sync.Once

// GetMessageController 获取消息 Controller 单例
func GetMessageController() *MessageController {
	messageControllerOnce.Do(func() {
		messageController = &MessageController{apiMessage: api.GetApiMessage()}
	})
	return messageController
}

// Router 注册路由（均需登录）
func (c *MessageController) Router(router *gin.Engine) {
	group := router.Group("/warm-nest/v1/message")
	group.GET("/list", web.Handler(c.apiMessage.List, widgets.Session))
	group.POST("/read", web.Handler(c.apiMessage.MarkRead, widgets.Session))
}
