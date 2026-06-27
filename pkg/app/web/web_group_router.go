// Package web @Author larry
// @Date 2025/4/14 10:58
// @Desc

package web

import (
	"github.com/gin-gonic/gin"
)

type GroupRouter interface {
	Router(group *gin.RouterGroup)
}
