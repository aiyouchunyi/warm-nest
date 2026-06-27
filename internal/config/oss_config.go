// Package config @Author larry
// @Date 2026/06/27
// @Desc 阿里云 OSS 配置（app.toml [oss] 段；AccessKey 由部署平台注入 env，占位空串）
//
// env 映射规则（framework configs.LoadEnv）：段名 oss + 字段 toml tag → 大写下划线 env 名，
// 如 accessKeyId → OSS_ACCESS_KEY_ID。故敏感值 toml 留空、由 deploy.sh 经 env-file 注入。
// 公共读 Bucket：上传走内网 endpoint 免流量费；返回给前端的 URL 用外网 publicEndpoint。

package config

import (
	"sync"

	"warm-nest/pkg/app/configs"
)

// OssConfig 阿里云 OSS 配置（[oss] 段）
type OssConfig struct {
	Endpoint        string `toml:"endpoint"`        // 内网 Endpoint（上传用，如 oss-cn-chengdu-internal.aliyuncs.com）
	PublicEndpoint  string `toml:"publicEndpoint"`  // 外网 Endpoint（拼公网访问 URL 用，如 oss-cn-chengdu.aliyuncs.com）
	Bucket          string `toml:"bucket"`          // Bucket 名
	Region          string `toml:"region"`          // 地域（如 cn-chengdu，v2 SDK 需要）
	AccessKeyId     string `toml:"accessKeyId"`     // RAM 子账号 AK ID（env OSS_ACCESS_KEY_ID 注入）
	AccessKeySecret string `toml:"accessKeySecret"` // RAM 子账号 AK Secret（env OSS_ACCESS_KEY_SECRET 注入）
}

var ossConf OssConfig
var ossConfOnce sync.Once

// OssConf 获取 OSS 配置单例
func OssConf() OssConfig {
	ossConfOnce.Do(func() {
		ossConf = configs.MustLoad[OssConfig]("oss")
	})
	return ossConf
}
