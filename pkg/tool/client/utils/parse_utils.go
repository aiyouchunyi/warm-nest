// Package utils @Author larry
// File parse_inner_err.go
// @Date 2024/6/19 10:48:00
// @Desc 解析服务业务内部错误
package utils

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/client/model"
	"warm-nest/pkg/utils/reflects"
	"warm-nest/pkg/utils/transforms"
)

// ParseErr 解析内部错误
func ParseErr[R model.IRespCode](_ *resty.Client, response *resty.Response) error {
	var resp R
	if response.IsSuccess() && reflects.IsStruct(response.Result()) {
		resp = transforms.TransformTo[R](response.Result())
	} else if !response.IsSuccess() && reflects.IsStruct(response.Error()) {
		resp = transforms.TransformTo[R](response.Error())
	} else if response.Body() != nil && response.Result() == nil && response.Error() == nil && json.Valid(response.Body()) {
		resp = transforms.Unmarshal[R](string(response.Body()))
	} else {
		return nil
	}

	if resp.Success() {
		return nil
	}

	return &errors.CodeError{
		Code: resp.GetCode(),
		Msg:  resp.GetMessage(),
	}
}
