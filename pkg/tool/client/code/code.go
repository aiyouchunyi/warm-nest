// Package code @Author larry
// File code.go
// @Date 2024/5/13 18:17:00
// @Desc 客户端错误码
package code

import (
	"warm-nest/pkg/app/errors"
)

var (
	TooManyRequests      = errors.NewCode("00000600", "请求过于频繁! url:[%s]")
	HttpResponseFailed   = errors.NewCode("00000601", "Http返回失败状态! status:[%d] body:[%s]")
	HttpResponseUnknown  = errors.NewCode("00000602", "Http返回未知状态! status:[%d] body:[%s]")
	HttpResponseParseErr = errors.NewCode("00000603", "Http返回值解析错误! body:[%v]")
	SignAccountAbsent    = errors.NewCode("00000620", "签名账户信息缺失! 账户:[%s]")
)
