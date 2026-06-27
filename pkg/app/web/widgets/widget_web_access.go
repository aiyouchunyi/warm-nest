// Package web @Author Larry
// @Date 2024/10/12 15:15
// @Desc

package widgets

import (
	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/app/web/config"
	"warm-nest/pkg/app/web/context"
	"warm-nest/pkg/app/web/header"
	"warm-nest/pkg/tool/auth"
	"warm-nest/pkg/utils/strings"
)

// Access 访问检查
func Access(c *gin.Context) error {
	if configs.IsLocal() || !config.ServerConf().AuthEnabled {
		return nil
	}
	appId := c.Request.Header.Get(header.App)
	if strings.NotBlank(appId) {
		return nil
	}
	webContext := context.GetContext(c)
	if webContext.TokenErr != nil {
		return webContext.TokenErr
	}

	return auth.GetAuthorizeService().Access(
		webContext.ReqUser,
		c.Request.Method,
		c.Request.URL.Path,
		c.Request.URL.RawQuery,
	)
}
