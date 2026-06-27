// Package validators @Author larry
// @Date 2024/4/16 11:15:00
// @Desc 参数校验器
package validators

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"

	"warm-nest/pkg/utils/reflects"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate
var validateOnce sync.Once

var trans, _ = ut.New(zh.New(), zh.New()).GetTranslator("zh")

// GetValidator 获取validator单例
func GetValidator() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		// 注册自定义验证函数
		for _, v := range validateFuncs {
			err := validate.RegisterValidation(v.Name, v.Validation)
			if err != nil {
				panic(fmt.Sprintf("注册验证函数失败: %s, 错误: %v", v.Name, err))
			}
		}
		// 注册中文翻译
		err := zh_translations.RegisterDefaultTranslations(validate, trans)
		if err != nil {
			panic(fmt.Sprintf("注册翻译器失败:%v", err))
		}

		// 注册自定义翻译函数
		for _, validateFunc := range validateFuncs {
			err2 := validate.RegisterTranslation(validateFunc.Name, trans, func(ut ut.Translator) error {
				return ut.Add(validateFunc.Name, validateFunc.Translation, true)
			}, func(ut ut.Translator, fe validator.FieldError) string {
				result, err2 := ut.T(validateFunc.Name, fe.Field(), fe.Param())
				if err2 != nil {
					result = strings.ReplaceAll(validateFunc.Translation, "{0}", fe.Field())
					result = strings.ReplaceAll(result, "{1}", fe.Param())
				}
				result = result + fmt.Sprintf(" 当前值:%v", fe.Value())
				return result
			})
			if err2 != nil {
				panic(fmt.Sprintf("注册翻译函数失败: %s, 错误: %v", validateFunc.Name, err))
			}
		}

	})
	return validate
}

func Validate(s interface{}) error {
	if !reflects.IsStruct(s) {
		return nil
	}
	err := GetValidator().Struct(s)
	// 如果为nil 说明校验通过
	if err == nil {
		return nil
	}
	// 如果是输入参数无效，则直接返回输入参数错误
	invalid, ok := err.(*validator.InvalidValidationError)
	if ok {
		return fmt.Errorf("无效的输入参数: %v", invalid.Error())
	}
	// 断言是ValidationErrors
	validationErrs := err.(validator.ValidationErrors)
	msgs := make([]string, 0)
	for _, validationErr := range validationErrs {
		msgs = append(msgs, validationErr.Translate(trans))
	}
	return fmt.Errorf("参数校验失败: %s", strings.Join(msgs, ", "))
}
