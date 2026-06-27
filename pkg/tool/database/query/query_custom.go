// Package query @Author larry
// File query_custom.go
// @Date 2024/6/6 16:21:00
// @Desc
package query

import (
	"fmt"
	"reflect"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/database/code"
)

// Custom 自定义条件
type Custom struct {
	Custom   string      `gorm:"comment:字段名" json:"custom" validate:"required"`
	Operator Operator    `gorm:"comment:操作符" json:"operator"`
	Value    interface{} `gorm:"comment:操作值" json:"value" `
}

func NewCustom(custom string, operator Operator, value ...interface{}) Custom {
	return Custom{
		Custom:   custom,
		Operator: operator,
		Value:    value,
	}
}

// ExpectSingle 期待数组
func (c *Custom) ExpectSingle() error {
	if c.Value == nil {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值不能为nil  %s:%v", c.Operator, c.Custom, c.Value))
	}
	if reflect.TypeOf(c.Value).Kind() != reflect.Slice {
		return nil
	}
	return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值必须是单值  %s:%v", c.Operator, c.Custom, c.Value))
}

// ExpectSlice 期待数组
func (c *Custom) ExpectSlice() error {
	if c.Value == nil {
		return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值不能为nil  %s:%v", c.Operator, c.Custom, c.Value))
	}
	if reflect.TypeOf(c.Value).Kind() == reflect.Slice {
		return nil
	}
	return errors.NewWithArgs(code.DBParamInvalid, fmt.Sprintf("%s操作符的值必须是数组  %s:%v", c.Operator, c.Custom, c.Value))
}
