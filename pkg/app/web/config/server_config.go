// Package config @Author larry
// @Date 2025/12/24 14:26
// @Desc

package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

type ServerConfig struct {
	Host            string `toml:"host" default:":8080"`         // 服务监听地址
	PprofEnabled    bool   `toml:"pprof_enabled" default:"true"` // 是否启用PPROF
	PprofPathPrefix string `toml:"pprof_path_prefix"`            // PPROF的路径前缀,
	AuthEnabled     bool   `toml:"auth_enabled" default:"false"` // 是否启用认证
}

var serverConf ServerConfig
var serverConfOnce sync.Once

// ServerConf 获取服务配置单例
func ServerConf() ServerConfig {
	serverConfOnce.Do(func() {
		serverConf = configs.MustLoad[ServerConfig](configs.WebServerConfPath)
	})
	return serverConf
}
