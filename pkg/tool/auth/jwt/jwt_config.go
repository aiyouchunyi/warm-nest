// Package jwt @Author larry
// File jwt_config.go
// @Date 2024/8/13 21:09:00
// @Desc
package jwt

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

type JWTConfig struct {
	SecretKey   string `toml:"secretKey" json:"secretKey" gorm:"comment:SecretKey"`
	PasswordKey string `toml:"passwordKey" json:"passwordKey" gorm:"comment:PasswordKey"`
	Expire      int64  `toml:"expire" json:"expire" gorm:"comment:过期时间[s]" default:"86400"`
}

var jwtConf JWTConfig
var jwtConfOnce sync.Once

// JWTConf 获取JWT配置单例
func JWTConf() JWTConfig {
	jwtConfOnce.Do(func() {
		jwtConf = configs.MustLoad[JWTConfig](configs.JWTConfPath)
	})
	return jwtConf
}
