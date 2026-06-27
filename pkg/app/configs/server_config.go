// Package configs @Author larry
// File server_config.go
// @Date 2024/5/8 14:33:00
// @Desc
package configs

import (
	"strings"
	"sync"

	"github.com/creasty/defaults"

	"warm-nest/pkg/app/configs/loader"
)

type ServerConfig struct {
	Name      string `toml:"name"`
	Site      string `toml:"site" default:"GB"`
	Mode      string `toml:"mode" default:"local"`
	SecretId  string `toml:"secretId"`
	SecretKey string `toml:"secretKey" default:"warm-nest"`
	IamEnable bool   `toml:"iamEnable" default:"true"`
}

var serverConf ServerConfig
var serverConfOnce sync.Once

// ServerConf 获取服务配置
func ServerConf() ServerConfig {
	serverConfOnce.Do(func() {
		_ = defaults.Set(&serverConf)
		err := loader.LoadTo(AppToml, ServerConfPath, &serverConf)
		if err != nil {
			panic(err)
		}
		err = loader.LoadEnv(ServerConfPath, &serverConf)
		if err != nil {
			panic(err)
		}
		if serverConf.Name == "" {
			panic("server name is empty")
		}
	})
	return serverConf
}

func GetMode() string {
	mode := ServerConf().Mode
	if mode == "" {
		return LocalMode
	}
	return mode
}

func GetSite() string {
	return ServerConf().Site
}

func GetAppName() string {
	return ServerConf().Name
}

func GetUpperMode() string {
	return strings.ToUpper(ServerConf().Mode)
}

func IsLocal() bool {
	return strings.EqualFold(GetMode(), LocalMode)
}

func IsProd() bool {
	return strings.EqualFold(GetMode(), ProdMode)
}

func GetSecretId() string {
	return ServerConf().SecretId
}

func GetSecretKey() string {
	return ServerConf().SecretKey
}

func IsIAMEnabled() bool {
	return ServerConf().IamEnable
}
