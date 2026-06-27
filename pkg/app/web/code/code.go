// Package code @Author larry
// File code.go
// @Date 2024/6/19 09:22:00
// @Desc web校验错误码
package code

import (
	"warm-nest/pkg/app/errors"
)

var (
	WebSystemInnerErr = errors.NewCode("00000700", "Web系统内部错误")
	WebCallParmaErr   = errors.NewCode("00000701", "Web参数错误! %s %s")
	WebRateLimit      = errors.NewCode("00000702", "Web请求频率过快")
)
