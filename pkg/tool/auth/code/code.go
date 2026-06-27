// Package code @Author larry
// File code.go
// @Date 2024/8/14 10:23:00
// @Desc
package code

import (
	"warm-nest/pkg/app/errors"
)

var (
	TokenInvalid     = errors.NewCode("00000800", "token无效")
	TokenGenFailed   = errors.NewCode("00000801", "token生成失败")
	TokenExpired     = errors.NewCode("00000802", "token已过期")
	LoginFailed      = errors.NewCode("00000820", "登入失败，请重试")
	UserDisabled     = errors.NewCode("00000821", "用户已被禁用")
	PermissionDenied = errors.NewCode("00000822", "接口权限不足!%s")
	PasswdError      = errors.NewCode("00000823", "密码错误")
	DefaultPassword  = errors.NewCode("00000824", "默认密码不可用")
)
