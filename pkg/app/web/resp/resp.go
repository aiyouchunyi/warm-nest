// Package resp @Author larry
// @Date 2024/4/15 18:24:00
// @Desc 响应处理
package resp

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/app/web/code"
	"warm-nest/pkg/utils/slices"
)

type Resp struct {
	Result  bool        `json:"result"`
	MCode   string      `json:"mcode,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

// New 创建响应
func New(code string, msg string, data interface{}) *Resp {
	return &Resp{
		Result:  code == "",
		MCode:   code,
		Message: msg,
		Data:    data,
	}
}

// NewSuccess 创建成功响应
func NewSuccess(data interface{}) *Resp {
	return New("", "success", data)
}

var IgnoreCode = []string{
	"00000800", "00000801", "00000802", "00000820", "00000821", "00000822", "00000823", "00000824",
}

// NewFail 创建失败响应
func NewFail(err error) *Resp {
	if err == nil {
		return New("", "success", nil)
	}
	var ce *errors.CodeError
	if errors.As(err, &ce) {
		if !slices.Contain(IgnoreCode, ce.Code) {
			logrus.WithError(err).Error("[WEB-FAIL]")
		}
		return New(ce.Code, ce.Msg, nil)
	}
	return New(code.WebSystemInnerErr.Code, err.Error(), nil)
}
