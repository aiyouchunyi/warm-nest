// Package web @Author larry
// @Date 2024/4/15 18:04:00
// @Desc controller接口
package web

import "github.com/gin-gonic/gin"

type Controller interface {
	Router(router *gin.Engine)
}
