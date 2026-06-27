// Package utils @Author larry
// @Date 2025/8/21 11:10
// @Desc

package utils

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func UrlInfo(c *gin.Context) string {
	return fmt.Sprintf("%s_%s", c.Request.Method, c.Request.URL.String())
}
