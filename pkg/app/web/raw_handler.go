// Package web @Author larry
// @Date 2026/3/30 20:12
// @Desc

package web

import (
	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web/resp"
	"warm-nest/pkg/app/web/widgets"
)

func RawHandler(f func(c *gin.Context) error, widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 执行处理函数
		for _, w := range widgets {
			if err := w(c); err != nil {
				resp.Fail(c, err)
				return
			}
		}
		err := f(c)
		if err != nil {
			resp.Fail(c, err)
			return
		}
	}
}
