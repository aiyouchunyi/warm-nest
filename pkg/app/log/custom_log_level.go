// Package log @Author larry
// File custom_log_level.go
// @Date 2024/8/12 16:45:00
// @Desc
package log

type LogLevel string

const (
	IGNORE LogLevel = "ignore"
	SIMPLE LogLevel = "simple"
	DETAIL LogLevel = "detail"
	DEBUG  LogLevel = "debug"
)
