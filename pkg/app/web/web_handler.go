// Package web @Author larry
// @Date 2024/4/16 15:36:00
// @Desc web处理器
package web

import (
	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/web/context"
	"warm-nest/pkg/app/web/resp"
	"warm-nest/pkg/app/web/utils"
	"warm-nest/pkg/app/web/validators"
	"warm-nest/pkg/app/web/widgets"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/transforms"
)

// NoRespHandler 有参数无返回值
func NoRespHandler[T any](f func(req T) error, widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return Handler(func(req T) (interface{}, error) {
		return nil, f(req)
	}, widgets...)
}

// Handler 有参数有返回值
func Handler[T any](f func(req T) (interface{}, error), widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 执行处理函数
		for _, w := range widgets {
			if err := w(c); err != nil {
				resp.Fail(c, err)
				return
			}
		}

		// 反序列化
		model, err := utils.UmMarshal[T](c)
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

			// 设置 Session
			setSession(c, &model)

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

// WithoutHandler 无参数无返回值
func WithoutHandler(f func() error, widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return NoReqHandler(func() (interface{}, error) {
		return nil, f()
	}, widgets...)
}

// NilHandler 无参数无返回值
func NilHandler(f func(), widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return NoReqHandler(func() (interface{}, error) {
		f()
		return nil, nil
	}, widgets...)
}

func NilRespHandler[T any](f func(req T), widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return NoRespHandler(func(req T) error {
		f(req)
		return nil
	}, widgets...)
}

// NoReqHandler 无参数有返回值
func NoReqHandler(f func() (interface{}, error), widgets ...widgets.WidgetFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		// 执行处理函数
		for _, w := range widgets {
			if err := w(c); err != nil {
				resp.Fail(c, err)
				return
			}
		}
		r, err := f()
		resp.Respond(c, r, err)
	}
}

// 获取请求用户
func setSession[T any](c *gin.Context, req *T) {
	if req == nil || !reflects.IsStruct(*req) {
		return
	}

	transforms.TransformIgnoreNil(context.GetContext(c), req)
}
