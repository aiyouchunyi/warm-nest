// Package strings @Author larry
// File string.go
// @Date 2024/5/28 10:03:00
// @Desc 字符串工具类
package strings

import (
	"strings"
)

func OrElse(s1 string, s2 string) string {
	if IsBlank(s1) {
		return s2
	}
	return s1
}

func EqualIgnoreNil(s1 string, s2 *string) bool {
	if s2 == nil {
		return true
	}
	return s1 == *s2
}

func ContainIgnoreNil(s1 string, s2 *string) bool {
	if s2 == nil {
		return true
	}
	return strings.Contains(s1, *s2)
}

func GetLastDot(s string) string {
	return GetLast(s, ".")
}

func GetLast(s string, sep string) string {
	parts := strings.Split(s, sep)
	return parts[len(parts)-1]
}

// IsBlank 判断字符串是否为空
func IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}

func NotBlank(str string) bool {
	return !IsBlank(str)
}

// BothBlank 判断多个字符串是否为空
func BothBlank(strs ...string) bool {
	for _, str := range strs {
		if !IsBlank(str) {
			return false
		}
	}
	return true
}

func HasBlank(strs ...string) bool {
	for _, str := range strs {
		if IsBlank(str) {
			return true
		}
	}
	return false
}

func RemoveBlank(strs ...string) []string {
	var result []string
	for _, str := range strs {
		if !IsBlank(str) {
			result = append(result, str)
		}
	}
	return result
}

func Split(s string, sep string) []string {
	if IsBlank(s) {
		return []string{}
	}
	return strings.Split(s, sep)
}

// Filters 过滤符串中的空格
func Filters(elms ...string) []string {
	var result []string
	for _, elem := range elms {
		if !IsBlank(elem) {
			result = append(result, elem)
		}
	}
	return result
}

func FiltersV2(elms []string) []string {
	return Filters(elms...)
}

func Join(sep string, elms ...string) string {
	return strings.Join(Filters(elms...), sep)
}

func JoinV2(sep string, elms []string) string {
	return strings.Join(Filters(elms...), sep)
}

func HasSuffix(s string, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func LowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(string(s[0])) + s[1:]
}
