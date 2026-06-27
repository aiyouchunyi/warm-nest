// Package config @Author larry
// @Date 2026/06/15
// @Desc 图片存储配置

package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

// 存储类型
const (
	StorageKindLocal = "LOCAL" // 本地文件（1 期）
	StorageKindOss   = "OSS"   // 对象存储（后期）
)

// StorageConfig 图片存储配置（app.toml [storage] 段）
type StorageConfig struct {
	Kind        string `toml:"kind" default:"LOCAL"`           // LOCAL|OSS
	BasePath    string `toml:"basePath"`                       // 本地存储根目录
	BaseUrl     string `toml:"baseUrl"`                        // 对外访问基础 URL
	MaxFileSize int64  `toml:"maxFileSize" default:"10485760"` // 单文件上限，默认 10MB
}

var storageConf StorageConfig
var storageConfOnce sync.Once

// StorageConf 获取存储配置单例
func StorageConf() StorageConfig {
	storageConfOnce.Do(func() {
		storageConf = configs.MustLoad[StorageConfig]("storage")
	})
	return storageConf
}
