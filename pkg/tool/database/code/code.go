// Package code @Author larry
// File code.go
// @Date 2024/5/22 14:12:00
// @Desc 数据库错误码
package code

import (
	"warm-nest/pkg/app/errors"
)

var (
	DBParamInvalid = errors.NewCode("00000301", "查询参数错误! 错误:%v")
	DBNotFound     = errors.NewCode("00000308", "未查询到数据! 条件:%v")
	DBAbnormal     = errors.NewCode("00000309", "数据库异常! 错误:%v")
	DBAffectZero   = errors.NewCode("00000310", "数据库操作影响行数为0! 条件:%v")
)
