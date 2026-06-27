// Package utils @Author larry
// @Date 2025/10/17 10:00
// @Desc

package utils

import (
	"strings"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/utils/transforms"
)

func PraseReq[T any](c *gin.Context) (T, error) {
	req, err := UmMarshal[T](c)
	if err != nil {
		return req, err
	}
	value := c.GetHeader("X-Feign-Key")
	if value == "" {
		return req, nil
	}
	xFeign := transforms.Unmarshal[session.XFeign](value)
	xFeign.UserGroup = c.GetHeader("X-User-Group")
	if xFeign.UserGroup == "picol" {
		xFeign.UserGroup = strings.ToUpper(xFeign.UserGroup)
	}
	if xFeign.UserGroup == "default" {
		xFeign.UserGroup = ""
	}
	transforms.TransformIgnoreNil(xFeign, &req)

	return req, nil
}
