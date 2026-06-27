// Package reflects @Author larry
// @Date 2024/4/12 17:37:00
// @Desc 自定义反射工具
package reflects

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

// IsNil 判断一个值是否为 nil
func IsNil(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	return (v.Kind() == reflect.Ptr ||
		v.Kind() == reflect.Slice ||
		v.Kind() == reflect.Map ||
		v.Kind() == reflect.Chan ||
		v.Kind() == reflect.Func) &&
		v.IsNil()
}

// IsStruct 判断一个值是否为结构体，考虑结构体指针
func IsStruct(i any) bool {
	if IsNil(i) {
		return false
	}
	v := reflect.ValueOf(i)
	kind := v.Kind()
	if kind == reflect.Ptr {
		if !v.IsNil() {
			return v.Elem().Kind() == reflect.Struct
		}
		return false
	}
	return kind == reflect.Struct
}

func IsSlice(i any) bool {
	if IsNil(i) {
		return false
	}
	v := reflect.ValueOf(i)
	kind := v.Kind()
	if kind == reflect.Ptr {
		if !v.IsNil() {
			return v.Elem().Kind() == reflect.Slice
		}
		return false
	}
	return kind == reflect.Slice
}

func IsArray(i any) bool {
	if IsNil(i) {
		return false
	}
	v := reflect.ValueOf(i)
	kind := v.Kind()
	if kind == reflect.Ptr {
		if !v.IsNil() {
			return v.Elem().Kind() == reflect.Array
		}
		return false
	}
	return kind == reflect.Array
}

// IsMap 判断一个值是否为Map，考虑Map指针
func IsMap(i any) bool {
	if IsNil(i) {
		return false
	}
	v := reflect.ValueOf(i)
	kind := v.Kind()
	if kind == reflect.Ptr {
		if !v.IsNil() {
			return v.Elem().Kind() == reflect.Map
		}
		return false
	}
	return kind == reflect.Map
}

func IsComplex(i any) bool {
	if IsNil(i) {
		return false
	}
	return IsStruct(i) || IsArray(i) || IsSlice(i) || IsMap(i)
}

// IsString 判断一个值是否为字符串
func IsString(i any) bool {
	if i == nil {
		return false
	}
	return reflect.ValueOf(i).Kind() == reflect.String
}

// New 创建一个安全的Map或Slice
func New[V any]() V {
	var v V
	valueType := reflect.TypeOf(v)
	if valueType == nil {
		return v
	}
	result := reflect.ValueOf(&v).Elem()
	switch valueType.Kind() {
	case reflect.Map:
		result.Set(reflect.MakeMap(valueType))
	case reflect.Slice:
		result.Set(reflect.MakeSlice(valueType, 0, 0))
	case reflect.Ptr:
		elem := valueType.Elem()
		switch elem.Kind() {
		case reflect.Struct:
			result.Set(reflect.New(elem))
		case reflect.Map:
			ptr := reflect.New(elem)
			ptr.Elem().Set(reflect.MakeMap(elem))
			result.Set(ptr)
		case reflect.Slice:
			ptr := reflect.New(elem)
			ptr.Elem().Set(reflect.MakeSlice(elem, 0, 0))
			result.Set(ptr)
		default:
			return v
		}
	default:
		return v
	}
	return v
}

func ToSlice(i any) []any {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Slice {
		return nil
	}
	s := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		s[i] = v.Index(i).Interface()
	}
	return s
}

func ToString(value any) string {
	// 特殊处理 decimal.Decimal 类型（在 IsComplex 之前）
	if dec, ok := value.(decimal.Decimal); ok {
		return dec.String()
	}

	if IsComplex(value) {
		bytes, err := safeJSONMarshal(value)
		if err != nil {
			logrus.WithError(err).Error("Marshal to json string failed!")
			return "<error>"
		}
		return string(bytes)
	}

	// 使用反射查看实际类型和值
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Ptr:
		// 如果是指针，获取指针指向的值（考虑指针可能为空）
		if !v.IsNil() {
			return ToString(v.Elem().Interface())
		}
		return "<nil>"
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		bytes, err := safeJSONMarshal(v)
		if err != nil {
			logrus.WithError(err).Error("Marshal to json string failed!")
			return "<error>"
		}
		return string(bytes)
	}
}

// MustFrom 从字符串转换为指定类型，如果失败则返回默认值
func MustFrom[V any](str string) V {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"str":   str,
				"type":  fmt.Sprintf("%T", *new(V)),
				"panic": r,
			}).Error("MustFrom panic recovered!")
		}
	}()

	v, err := From[V](str)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"str":  str,
			"type": fmt.Sprintf("%T", v),
		}).WithError(err).Warn("Failed to convert string to type, returning zero value")
		return v
	}
	return v
}

// FromWithErr 从字符串转换为指定类型
func FromWithErr[V any](str string, err error) (V, error) {
	var result V
	if err != nil {
		return result, err
	}
	return From[V](str)
}

// From 从字符串转换为指定类型
func From[V any](str string) (V, error) {
	var result V
	rv := reflect.ValueOf(&result).Elem()
	if str == "" {
		return result, nil
	}

	// 先处理特殊类型（在 switch 之前）
	// 处理 time.Time
	if reflect.TypeOf(result) == reflect.TypeOf(time.Time{}) {
		parsed, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return result, fmt.Errorf("parse time failed: %w", err)
		}
		rv.Set(reflect.ValueOf(parsed))
		return result, nil
	}

	// 处理 decimal.Decimal
	if reflect.TypeOf(result) == reflect.TypeOf(decimal.Decimal{}) {
		parsed, err := decimal.NewFromString(str)
		if err != nil {
			return result, fmt.Errorf("parse decimal failed: %w", err)
		}
		rv.Set(reflect.ValueOf(parsed))
		return result, nil
	}

	// 处理 sql.NullString, sql.NullInt64 等（str 已经在函数开头检查过不为空）
	switch reflect.TypeOf(result) {
	case reflect.TypeOf(sql.NullString{}):
		rv.Set(reflect.ValueOf(sql.NullString{String: str, Valid: true}))
		return result, nil
	case reflect.TypeOf(sql.NullInt64{}):
		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return result, fmt.Errorf("parse null int64 failed: %w", err)
		}
		rv.Set(reflect.ValueOf(sql.NullInt64{Int64: parsed, Valid: true}))
		return result, nil
	case reflect.TypeOf(sql.NullFloat64{}):
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return result, fmt.Errorf("parse null float64 failed: %w", err)
		}
		rv.Set(reflect.ValueOf(sql.NullFloat64{Float64: parsed, Valid: true}))
		return result, nil
	case reflect.TypeOf(sql.NullBool{}):
		parsed, err := strconv.ParseBool(str)
		if err != nil {
			return result, fmt.Errorf("parse null bool failed: %w", err)
		}
		rv.Set(reflect.ValueOf(sql.NullBool{Bool: parsed, Valid: true}))
		return result, nil
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return result, err
		}
		rv.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return result, fmt.Errorf("parse uint failed: %w", err)
		}
		rv.SetUint(parsed)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return result, err
		}
		rv.SetFloat(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(str)
		if err != nil {
			return result, err
		}
		rv.SetBool(parsed)
	case reflect.Struct, reflect.Slice, reflect.Map:
		// 如果是结构体，尝试用 JSON 解析字符串
		err := safeJSONUnmarshal([]byte(str), &result)
		if err != nil {
			return result, err
		}
	default:
		err := fmt.Errorf("unsupported target type: %s", rv.Kind())
		return result, err
	}
	return result, nil
}

func MustCast[V any](v any) V {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"value": v,
				"type":  fmt.Sprintf("%T", *new(V)),
				"panic": r,
			}).Error("MustCast panic recovered!")
		}
	}()

	result, err := Cast[V](v)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"value": v,
			"type":  fmt.Sprintf("%T", result),
		}).WithError(err).Warn("Failed to cast value to type, returning zero value")
		return result
	}
	return result
}

func CastWithErr[V any](v any, err error) (V, error) {
	var result V
	if err != nil {
		return result, err
	}
	return Cast[V](v)
}

func Cast[V any](v any) (result V, err error) {
	// 添加 panic recovery
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cast panic: %v (value type: %T, target type: %T)", r, v, result)
			logrus.WithFields(logrus.Fields{
				"value":      v,
				"targetType": fmt.Sprintf("%T", result),
				"panic":      r,
			}).Error("Cast panic recovered!")
		}
	}()

	if v == nil {
		return result, nil
	}
	typeV := reflect.TypeOf(result)
	if typeV == nil {
		return result, nil
	}
	// 1. 类型完全匹配
	if reflect.TypeOf(v) == typeV {
		return v.(V), nil
	}
	// 2. 指针类型自动解引用
	if typeV.Kind() == reflect.Ptr && reflect.TypeOf(v) == typeV.Elem() {
		ptr := reflect.New(typeV.Elem())
		ptr.Elem().Set(reflect.ValueOf(v))
		return ptr.Interface().(V), nil
	}
	// 3. 字符串转目标类型
	if str, ok := v.(string); ok {
		return From[V](str)
	}
	// 4. []byte 或 json.RawMessage
	if b, ok := v.([]byte); ok {
		if typeV.Kind() == reflect.String {
			return any(string(b)).(V), nil
		}
		err := safeJSONUnmarshal(b, &result)
		return result, err
	}
	if raw, ok := v.(json.RawMessage); ok {
		if typeV.Kind() == reflect.String {
			return any(string(raw)).(V), nil
		}
		err := safeJSONUnmarshal(raw, &result)
		return result, err
	}
	// 5. 基本类型转字符串再转目标类型
	switch val := v.(type) {
	case int, int8, int16, int32, int64, float32, float64, bool, uint, uint8, uint16, uint32, uint64:
		return From[V](fmt.Sprintf("%v", val))
	}
	// 6. map/slice/struct 用json序列化再反序列化
	kind := reflect.TypeOf(v).Kind()
	if kind == reflect.Map || kind == reflect.Slice || kind == reflect.Struct {
		b, err := safeJSONMarshal(v)
		if err != nil {
			return result, err
		}
		err = safeJSONUnmarshal(b, &result)
		return result, err
	}
	return result, fmt.Errorf("cannot convert %T to %T", v, result)
}

// SetFieldValue 根据字段类型将字符串赋值给字段
func SetFieldValue(fv reflect.Value, val string) error {
	if !fv.CanSet() {
		return fmt.Errorf("field is not settable: type=%s", fv.Type().String())
	}
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int failed: %w", err)
		}
		fv.SetInt(iv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uv, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("parse uint failed: %w", err)
		}
		fv.SetUint(uv)
	case reflect.Float32, reflect.Float64:
		fv2, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("parse float failed: %w", err)
		}
		fv.SetFloat(fv2)
	case reflect.Bool:
		bv, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("parse bool failed: %w", err)
		}
		fv.SetBool(bv)
	default:
		return fmt.Errorf("unsupported field type: %s", fv.Kind())
	}
	return nil
}

// safeJSONMarshal JSON 序列化，捕获 panic
func safeJSONMarshal(v any) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"value": fmt.Sprintf("%#v", v),
				"panic": r,
			}).Error("JSON Marshal panic recovered!")
		}
	}()
	return json.Marshal(v)
}

// safeJSONUnmarshal JSON 反序列化，捕获 panic
func safeJSONUnmarshal(data []byte, v any) error {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithFields(logrus.Fields{
				"data":  string(data),
				"panic": r,
			}).Error("JSON Unmarshal panic recovered!")
		}
	}()
	return json.Unmarshal(data, v)
}
