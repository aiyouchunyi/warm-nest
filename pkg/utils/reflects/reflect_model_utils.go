// Package reflects @Author larry
// @Date 2025/3/21 19:20
// @Desc

package reflects

import (
	"errors"
	"fmt"
	"reflect"
)

func GetModelField(model any, fieldName string) (any, error) {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if !val.IsValid() {
		return nil, fmt.Errorf("invalid model: %T", model)
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found in model %T", fieldName, model)
	}
	return field.Interface(), nil
}

// ExistsModelField 检查模型是否存在字段
func ExistsModelField(model any, fieldName string) bool {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if !val.IsValid() {
		return false
	}

	field := val.FieldByName(fieldName)
	return field.IsValid()
}

// SetModelField 使用反射设置模型的字段
func SetModelField(model any, fieldName string, value any) error {
	v := reflect.ValueOf(model)

	// 确保传入的是指针
	if v.Kind() != reflect.Ptr {
		return errors.New("model must be a pointer")
	}

	// 获取指针指向的值
	v = v.Elem()

	// 确保模型是结构体
	if v.Kind() != reflect.Struct {
		return errors.New("model must be a struct")
	}

	// 获取字段
	field := v.FieldByName(fieldName)

	// 检查字段是否存在
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s in model", fieldName)
	}

	// 检查字段是否可设置
	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}

	// 将值设置到字段
	val := reflect.ValueOf(value)

	// 类型检查和转换，如果需要
	if field.Type() != val.Type() {
		if val.Type().ConvertibleTo(field.Type()) {
			val = val.Convert(field.Type())
		} else {
			return fmt.Errorf("provided value type (%v) didn't match object field type (%v)", val.Type(), field.Type())
		}
	}

	field.Set(val)
	return nil
}

// ModelName 获取模型名称
func ModelName(model any) string {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val.Type().Name()
}

func ModelNameV2[T any]() string {
	var model T
	return ModelName(model)
}
