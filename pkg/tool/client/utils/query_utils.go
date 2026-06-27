package utils

import (
	"reflect"
	"strconv"
	"strings"
)

// ToQuery 把 struct 反射成 query 参数 map。
//
// tag 优先级：form > json，二者皆缺则该字段跳过。
// 零值（"" / 0 / false / nil ptr）一律跳过，避免无意义参数。
func ToQuery(data interface{}) map[string]string {
	queries := make(map[string]string)
	if data == nil {
		return queries
	}
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return queries
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldT := t.Field(i)
		if !fieldT.IsExported() {
			continue
		}
		tag := queryTagName(fieldT)
		if tag == "" {
			continue
		}
		value := getFieldValue(v.Field(i))
		if value == "" {
			continue
		}
		queries[tag] = value
	}
	return queries
}

// queryTagName 返回字段的 query key：form 优先，json 兑底。
//
// tag 形如 `form:"name,omitempty"`，按逗号取首段。值为 "-" 视为忽略。
func queryTagName(field reflect.StructField) string {
	for _, key := range []string{"form", "json"} {
		raw := field.Tag.Get(key)
		if raw == "" {
			continue
		}
		name := raw
		if idx := strings.Index(raw, ","); idx >= 0 {
			name = raw[:idx]
		}
		if name == "-" {
			return ""
		}
		if name != "" {
			return name
		}
	}
	return ""
}

// getFieldValue 把字段值序列化为 query string；零值返回 ""。
//
// Float 用 'f' -1 精度，避免 silent truncation；上游需要定点格式自行约束。
func getFieldValue(fieldV reflect.Value) string {
	switch fieldV.Kind() {
	case reflect.String:
		return fieldV.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fieldV.Int() == 0 {
			return ""
		}
		return strconv.FormatInt(fieldV.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if fieldV.Uint() == 0 {
			return ""
		}
		return strconv.FormatUint(fieldV.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		if fieldV.Float() == 0 {
			return ""
		}
		return strconv.FormatFloat(fieldV.Float(), 'f', -1, 64)
	case reflect.Bool:
		if !fieldV.Bool() {
			return ""
		}
		return "true"
	case reflect.Ptr:
		if fieldV.IsNil() {
			return ""
		}
		return getFieldValue(fieldV.Elem())
	}
	return ""
}
