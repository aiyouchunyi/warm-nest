// Package resp @Author larry
// @Date 2024/4/15 19:41:00
// @Desc web响应处理
package resp

import (
	"github.com/gin-gonic/gin"
)

func Respond(c *gin.Context, data interface{}, err error) {
	if err != nil {
		Fail(c, err)
		return
	}
	Success(c, data)
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, NewSuccess(data))
}

func Fail(c *gin.Context, err error) {
	c.JSON(200, NewFail(err))
}
