// Package errors @Author larry
// @Date 2024/4/11 15:41:00
// @Desc 编码错误(业务未重复错误码使用)
package errors

import (
	"errors"
	"fmt"
)

// CodeError 编码错误
type CodeError struct {
	Code string // 错误码
	Msg  string // 错误信息,可带格式化字符串
}

func NewCode(code string, message string) CodeError {
	return CodeError{
		Code: code,
		Msg:  message,
	}
}

func NewWithArgs(codeError CodeError, args ...interface{}) error {
	return &CodeError{
		Code: codeError.Code,
		Msg:  fmt.Sprintf(codeError.Msg, args...),
	}
}

func (c *CodeError) Error() string {
	return fmt.Sprintf("%s-%s", c.Code, c.Msg)
}

func (c *CodeError) Is(err error) bool {
	if err == nil {
		return false
	}
	var ce *CodeError
	if errors.As(err, &ce) {
		if ce.Code == c.Code {
			return true
		}
	}
	return false
}
