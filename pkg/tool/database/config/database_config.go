// Package config @Author larry
// @Date 2025/12/22 19:56
// @Desc

package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

type DatabaseConfig struct {
	Mysql string `toml:"mysql" secret:"mysql.url" json:"mysql"`
	Mongo string `toml:"mongo" secret:"mongo.url" json:"mongo"`
}

var databaseConf DatabaseConfig
var databaseConfOnce sync.Once

// DatabaseConf 获取数据库配置单例
func DatabaseConf() DatabaseConfig {
	databaseConfOnce.Do(func() {
		databaseConf = configs.MustLoadSecret[DatabaseConfig](configs.DatabaseConfPath)
	})
	return databaseConf
}
