// Package web @Author larry
// @Date 2026/3/30 00:00
// @Desc 用户请求处理器
package web

import (
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/web/resp"
	"warm-nest/pkg/app/web/utils"
	"warm-nest/pkg/app/web/validators"
	"warm-nest/pkg/app/web/widgets"
	"warm-nest/pkg/utils/reflects"
)

// NoRespUserHandler 有参数无返回值，自动处理 XFeign 头部信息注入（f: 业务处理函数, widgets: 拦截器列表, 返回: gin.HandlerFunc）
func NoRespUserHandler[T any](f func(req T) error, widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return UserHandler(func(req T) (interface{}, error) {
		return nil, f(req)
	}, widgets...)
}

// UserHandler 有参数有返回值，自动处理 XFeign 头部信息注入（f: 业务处理函数, widgets: 拦截器列表, 返回: gin.HandlerFunc）
func UserHandler[T any](f func(req T) (interface{}, error), widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 执行处理函数
		for _, w := range widgets {
			if err := w(c); err != nil {
				resp.Fail(c, err)
				return
			}
		}

		// 反序列化（使用 PraseReq 自动处理 XFeign）
		model, err := utils.PraseReq[T](c)
		if err != nil {
			resp.Fail(c, err)
			return
		}
		if reflects.IsStruct(model) {
			// 校验参数
			err = validators.Validate(model)
			if err != nil {
				logrus.WithError(err).Errorf("Validate failed!")
				resp.Fail(c, err)
				return
			}

			// 设置默认值
			err = defaults.Set(&model)
			if err != nil {
				resp.Fail(c, err)
				return
			}
		}

		r, err := f(model)
		resp.Respond(c, r, err)
	}
}

// NilRespUserHandler 有参数无返回值，自动处理 XFeign 头部信息注入（f: 业务处理函数, widgets: 拦截器列表, 返回: gin.HandlerFunc）
func NilRespUserHandler[T any](f func(req T), widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return NoRespUserHandler(func(req T) error {
		f(req)
		return nil
	}, widgets...)
}
