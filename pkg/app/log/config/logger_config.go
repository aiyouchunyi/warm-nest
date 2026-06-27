// Package log @Author larry
// @Date 2025/12/24 14:28
// @Desc

package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

type LoggerConfig struct {
	Format     string `toml:"format" default:"text"`
	Level      string `toml:"level" default:"info"`
	TimeFormat string `toml:"time_format" default:"20060102 15:04:05.000"`
}

var loggerConf LoggerConfig
var loggerConfOnce sync.Once

// LoggerConf 获取Logger配置单例
func LoggerConf() LoggerConfig {
	loggerConfOnce.Do(func() {
		loggerConf = configs.MustLoad[LoggerConfig](configs.LoggerConfPath)
	})
	return loggerConf
}
