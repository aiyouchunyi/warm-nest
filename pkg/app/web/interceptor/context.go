// Package interceptor @Author larry
// @Date 2025/5/27 10:36
// @Desc

package interceptor

import (
	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web/context"
	"warm-nest/pkg/app/web/header"
	"warm-nest/pkg/tool/auth"
)

func Context(authEnable bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL.String()
		if url == "/ping" {
			return
		}

		token := c.Request.Header.Get(header.Token)
		if token != "" && authEnable {
			ctx := context.GetContext(c)
			ctx.ReqUser, ctx.User, ctx.TokenErr = auth.GetAuthorizeService().Verify(token)
			context.SetContext(c, ctx)
		}
		c.Next()
	}
}
