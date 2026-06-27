// Package widgets @Author larry
// @Date 2025/2/6 15:34
// @Desc

package widgets

import (
	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/app/web/context"
	"warm-nest/pkg/app/web/header"
	"warm-nest/pkg/utils/strings"
)

// Session Session检查
func Session(c *gin.Context) error {
	if configs.IsLocal() {
		return nil
	}
	appId := c.Request.Header.Get(header.App)
	if strings.NotBlank(appId) {
		return nil
	}
	webContext := context.GetContext(c)
	return webContext.TokenErr
}
