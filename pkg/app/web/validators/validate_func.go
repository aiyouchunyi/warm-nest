// Package validators @Author larry
// File validate_func.go
// @Date 2024/4/28 16:50:00
// @Desc 自定义校验函数
package validators

import (
	"reflect"

	"github.com/go-playground/validator/v10"

	"warm-nest/pkg/utils/numbers"
)

var (
	DecimalStrFunc = "decimalStr"
	DecimalMaxFunc = "decimalMax"
	DecimalMinFunc = "decimalMin"
	DecimalGtFunc  = "decimalGt"
	DecimalNeFunc  = "decimalNe"
	DecimalEqFunc  = "decimalEq"
	DecimalLtFunc  = "decimalLt"
	ArrayStrFunc   = "arrayStr"
)

type ValidateFunc struct {
	Name        string                             // Name 返回校验函数名称
	Translation string                             // Translation 返回校验函数的翻译文本
	Validation  func(fl validator.FieldLevel) bool // Validation 校验函数
}

var validateFuncs = []ValidateFunc{
	{
		Name:        DecimalStrFunc,
		Translation: "{0}必须是数字字符串",
		Validation: func(fl validator.FieldLevel) bool {
			return numbers.Valid(GetValue(fl)) == nil
		},
	},
	{
		Name:        DecimalMaxFunc,
		Translation: "{0}不能大于{1}",
		Validation: func(fl validator.FieldLevel) bool {
			out, err := numbers.Compare(GetValue(fl), fl.Param())
			if err != nil {
				return false
			}
			return out <= 0
		},
	},
	{
		Name:        DecimalMinFunc,
		Translation: "{0}不能小于{1}",
		Validation: func(fl validator.FieldLevel) bool {
			out, err := numbers.Compare(GetValue(fl), fl.Param())
			if err != nil {
				return false
			}
			return out >= 0
		},
	},
	{
		Name:        DecimalGtFunc,
		Translation: "{0}必须大于{1}",
		Validation: func(fl validator.FieldLevel) bool {
			out, err := numbers.Compare(GetValue(fl), fl.Param())
			if err != nil {
				return false
			}
			return out > 0
		},
	},
	{
		Name:        DecimalLtFunc,
		Translation: "{0}必须小于{1}",
		Validation: func(fl validator.FieldLevel) bool {
			out, err := numbers.Compare(GetValue(fl), fl.Param())
			if err != nil {
				return false
			}
			return out < 0
		},
	},
	{
		Name:        DecimalNeFunc,
		Translation: "{0}不能等于{1}",
		Validation: func(fl validator.FieldLevel) bool {
			out, err := numbers.Compare(GetValue(fl), fl.Param())
			if err != nil {
				return false
			}
			return out != 0
		},
	},
	{
		Name:        DecimalEqFunc,
		Translation: "{0}必须等于{1}",
		Validation: func(fl validator.FieldLevel) bool {
			out, err := numbers.Compare(GetValue(fl), fl.Param())
			if err != nil {
				return false
			}
			return out == 0
		},
	},
	{
		Name:        ArrayStrFunc,
		Translation: "{0}必须是字符串数组",
		Validation: func(fl validator.FieldLevel) bool {
			if fl.Field().Kind() != reflect.Slice {
				return false
			}
			if fl.Field().Len() == 0 {
				return false
			}
			for i := 0; i < fl.Field().Len(); i++ {
				if fl.Field().Index(i).Kind() != reflect.String {
					return false
				}
				if fl.Field().Index(i).String() == "" {
					return false
				}
			}
			return true
		},
	},
}

// GetValue 获取字段值
func GetValue(fl validator.FieldLevel) string {
	if fl.Field().Kind() == reflect.Ptr {
		if fl.Field().IsNil() {
			return ""
		}
		return fl.Field().Elem().String()
	}
	return fl.Field().String()
}
