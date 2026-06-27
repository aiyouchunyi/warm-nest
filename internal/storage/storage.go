// Package storage @Author larry
// @Date 2026/06/15
// @Desc 图片存储抽象（接口 + 按配置选实现）
//
// 本地文件实现 + 阿里云 OSS 实现，改配置 kind=LOCAL|OSS 切换，业务零改。
// OSS 为公共读 Bucket：Save 返回完整永久可访问 URL，直接入库、展示无需签名。

package storage

import (
	"sync"

	"warm-nest/internal/config"
)

// Storage 图片存储接口
type Storage interface {
	// Save 保存字节流到 bizDir 下的 filename，返回对外可访问 URL（入库用）
	Save(data []byte, bizDir, filename string) (url string, err error)
	// Get 读取（bizDir/filename 相对路径）
	Get(relPath string) ([]byte, error)
}

var instance Storage
var once sync.Once

// Get 获取存储实现单例（按配置 kind 选）
func Get() Storage {
	once.Do(func() {
		switch config.StorageConf().Kind {
		case config.StorageKindOss:
			instance = newOssStorage()
		default:
			instance = newLocalStorage()
		}
	})
	return instance
}
