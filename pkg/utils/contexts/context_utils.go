// Package contexts @Author larry
// File context_utils.go
// @Date 2024/5/17 10:46:00
// @Desc
package contexts

import (
	"context"

	"warm-nest/pkg/utils/transforms"
)

// WithValue  设置值
func WithValue[T any](ctx context.Context, key string, options ...func(t *T)) context.Context {
	value, _ := GetValue[T](ctx, key)
	for _, option := range options {
		option(&value)
	}
	return SetValue(ctx, key, value)
}

// SetValue 设置Context
func SetValue(ctx context.Context, key string, value any) context.Context {
	return context.WithValue(ctx, key, transforms.Marshal(value))
}

// GetValue 获取值Context
func GetValue[T any](ctx context.Context, key string) (T, bool) {
	var t T
	value := ctx.Value(key)
	if value == nil {
		return t, false
	}
	return transforms.Unmarshal[T](value.(string)), true
}
