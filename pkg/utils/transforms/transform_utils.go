// Package transforms @Author larry
// @Date 2024/4/16 16:35:00
// @Desc 转换工具类
package transforms

import (
	"encoding/json"
	"reflect"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/reflects"
)

func Marshal(src any) string {
	if reflects.IsString(src) {
		return src.(string)
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"src": src,
		}).WithError(err).Error("Marshal to json string failed!")
	}
	return string(bytes)
}

func Unmarshal[V any](src string) V {
	var v = reflects.New[V]()
	if src == "" {
		return v
	}
	err := json.Unmarshal([]byte(src), &v)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"src": src,
		}).WithError(err).Error("unmarshal to struct failed!")
	}
	return v
}

func UnmarshalWithErr[V any](src string, err error) (V, error) {
	var v = reflects.New[V]()
	if err != nil {
		return v, err
	}
	return Unmarshal[V](src), nil
}

func UnmarshalTo(src string, to any) {
	if src == "" {
		return
	}
	err := json.Unmarshal([]byte(src), to)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"src": src,
			"to":  to,
		}).WithError(err).Error("unmarshal to struct failed!")
	}
}

func Unmarshals[V any](src string) []V {
	var v []V
	if src == "" {
		return v
	}
	err := json.Unmarshal([]byte(src), &v)
	if err != nil {
		panic("unmarshal to failed")
	}
	return v
}

// CopyTo 使用 JSON 编码/解码实现深复制
func CopyTo[T any](src any) T {
	bytes, err := json.Marshal(src)
	var dst T
	if err != nil {
		panic("call param err, src marshal failed")
	}

	err = json.Unmarshal(bytes, &dst)
	if err != nil {
		panic("call param err, dst unmarshal failed")
	}

	return dst
}

// TransformsTo 使用反射实现结构体字段赋值
// src: 源结构体
// dst: 目标结构体指针
func TransformsTo[S any, T any](srcs []S) []T {
	var dts = make([]T, 0, len(srcs))
	for _, src := range srcs {
		var dst T
		TransformIgnoreNil(src, &dst)
		dts = append(dts, dst)
	}
	return dts
}

// TransformTo 使用反射实现结构体字段赋值
// src: 源结构体
// dst: 目标结构体指针
func TransformTo[V any](src interface{}) V {
	var model V
	TransformIgnoreNil(src, &model)
	return model
}

// TransformToPtr 使用反射实现结构体字段赋值
// src: 源结构体
// dst: 目标结构体指针
func TransformToPtr[V any](src interface{}) *V {
	var model V
	TransformIgnoreNil(src, &model)
	return &model
}

// TransformIgnoreNil 使用反射实现结构体字段赋值
// src: 源结构体
// dst: 目标结构体指针
// ignoreNil: 是否忽略空指针
func TransformIgnoreNil(src interface{}, dst interface{}) {
	Transform(src, dst, true)
}

// Transform 使用反射实现结构体字段赋值
// src: 源结构体
// dst: 目标结构体指针
// ignoreNil: 是否忽略空指针
func Transform(src interface{}, dst interface{}, ignoreNil bool) {
	validateInputs(src, dst)

	srcType, srcValue := getTypeAndValue(src)
	dstType := reflect.TypeOf(dst)
	dstValue := reflect.ValueOf(dst)

	for i := 0; i < srcType.NumField(); i++ {
		srcField := srcType.Field(i)
		srcFieldValue := srcValue.Field(i)

		// 匿名字段
		if srcField.Anonymous {
			if reflects.IsNil(srcFieldValue) {
				continue
			}

			Transform(srcFieldValue.Interface(), dst, ignoreNil)
			continue
		}

		// 非匿名字段
		dstField, ok := dstType.Elem().FieldByName(srcField.Name)
		if !ok || (ignoreNil && reflects.IsNil(srcFieldValue)) {
			continue
		}

		dstFieldValue := dstValue.Elem().FieldByName(srcField.Name)
		if srcField.Type == dstField.Type {
			dstFieldValue.Set(srcFieldValue)
			continue
		}
		if srcFieldValue.Kind() == reflect.Ptr && dstFieldValue.Kind() != reflect.Ptr && !srcFieldValue.IsNil() && srcFieldValue.Elem().Type() == dstFieldValue.Type() {
			dstFieldValue.Set(srcFieldValue.Elem())
			continue
		}
		if srcFieldValue.Kind() != reflect.Ptr && dstFieldValue.Kind() == reflect.Ptr && dstFieldValue.CanAddr() && !dstFieldValue.IsNil() && dstFieldValue.Elem().Type() == srcFieldValue.Type() {
			dstFieldValue.Elem().Set(srcFieldValue)
			continue
		}
	}
}

func validateInputs(src interface{}, dst interface{}) {
	if reflects.IsNil(src) {
		panic("call param err, to is nil")
	}
	if reflects.IsNil(dst) {
		panic("call param err, to is nil")
	}
	srcType := reflect.TypeOf(src)
	if srcType.Kind() != reflect.Struct && srcType.Kind() == reflect.Ptr && srcType.Elem().Kind() != reflect.Struct {
		panic("call param err, src must be struct or struct pointer")
	}

	dstType := reflect.TypeOf(dst)
	if dstType.Kind() != reflect.Ptr || dstType.Elem().Kind() != reflect.Struct {
		panic("call param err, dst must be struct pointer")
	}
}

func getTypeAndValue(i interface{}) (reflect.Type, reflect.Value) {
	t := reflect.TypeOf(i)
	v := reflect.ValueOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	return t, v
}
