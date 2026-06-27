// Package widget @Author larry
// @Date 2025/2/5 19:24
// @Desc

package widgets

import (
	"github.com/gin-gonic/gin"
)

type WidgetFunc func(c *gin.Context) error
