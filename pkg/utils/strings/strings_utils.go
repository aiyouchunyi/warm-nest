// Package strings @Author larry
// @Date 2025/5/13 10:18
// @Desc

package strings

import (
	"strings"
	"unicode"
)

func TrimSpace(strs []string) []string {
	for i, str := range strs {
		strs[i] = strings.TrimSpace(str)
	}
	return strs
}

// CamelToUnderline 驼峰转下划线
func CamelToUnderline(s string) string {
	result := make([]rune, 0, len(s)*2)
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// CamelToSnake 驼峰转下划线（不处理 . 分隔符，由调用方负责）
func CamelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(rune(s[i-1])) || unicode.IsDigit(rune(s[i-1]))) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// SnakeToCamel 下划线转驼峰（首字母小写）
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 0 {
		return ""
	}
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		if i == 0 {
			parts[i] = strings.ToLower(parts[i])
			continue
		}
		runes := []rune(parts[i])
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, "")
}
