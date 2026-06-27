// Package context @Author Larry
// @Date 2024/10/12 14:53
// @Desc

package context

import (
	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web/header"
	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/traces"
	"warm-nest/pkg/utils/transforms"
)

const WebContextKey = "web_context"

type WebContext struct {
	ReqId    string       `gorm:"comment:请求唯一标识" json:"reqId"`
	ReqUser  string       `gorm:"comment:请求用户" json:"reqUser"`
	User     session.User `gorm:"comment:用户信息" json:"user"`
	TokenErr error        `gorm:"comment:token错误" json:"-"`
}

// GetContext 获取上下文
func GetContext(c *gin.Context) WebContext {
	if v, ok := c.Get(WebContextKey); ok {
		return v.(WebContext)
	}
	webContext := WebContext{
		ReqId:   strings.OrElse(c.GetHeader(header.Trace), traces.TraceId("W")),
		ReqUser: strings.OrElse(c.Request.Header.Get(header.User), strings.OrElse(c.Request.Header.Get(header.App), "system")),
	}
	webContextStr := c.Request.Header.Get(header.WebContext)
	if webContextStr != "" {
		webContext = transforms.Unmarshal[WebContext](webContextStr)
	}
	SetContext(c, webContext)
	return webContext
}

// SetContext 设置上下文
func SetContext(c *gin.Context, ctx WebContext) {
	c.Set(WebContextKey, ctx)
	c.Request.Header.Set(header.WebContext, transforms.Marshal(ctx))
}

