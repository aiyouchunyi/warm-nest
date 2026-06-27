// Package conds @Author larry
// @Date 2025/1/2 14:59
// @Desc

package conds

func Cond[T any](cond bool, left T, right T) T {
	if cond {
		return left
	}
	return right
}
