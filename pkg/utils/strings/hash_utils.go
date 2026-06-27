// Package utils @Author larry
// @Date 2025/6/6 19:09
// @Desc

package strings

import (
	"strings"
)

func HashEqual(a, b string) bool {
	if a == b {
		return true
	}
	if a == "" || b == "" {
		return false
	}
	if strings.EqualFold(a, b) {
		return true
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}
	return false
}
