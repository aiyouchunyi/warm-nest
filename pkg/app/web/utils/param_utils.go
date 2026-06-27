// Package utils @Author larry
// @Date 2024/4/15 18:28:00
// @Desc 工具类
package utils

import (
	errors2 "errors"
	"io"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/app/web/code"
)

func UmMarshal[T any](c *gin.Context) (T, error) {
	var model T
	modelName := reflect.TypeOf(model).Name()
	err := c.ShouldBindQuery(&model)
	if err != nil {
		logrus.WithError(err).Errorf("failed query parse to %s!", modelName)
		return model, errors.NewWithArgs(code.WebCallParmaErr, "failed query parse!", err.Error())
	}
	err = c.ShouldBindUri(&model)
	if err != nil {
		logrus.WithError(err).Errorf("failed uri parse to %s!", modelName)
		return model, errors.NewWithArgs(code.WebCallParmaErr, "failed uri parse!", err.Error())
	}
	err = c.ShouldBindJSON(&model)
	if err != nil && !errors2.Is(err, io.EOF) {
		logrus.WithError(err).Errorf("failed body parse to %s!", modelName)
		return model, errors.NewWithArgs(code.WebCallParmaErr, "failed body parse!", err.Error())
	}
	return model, nil
}
