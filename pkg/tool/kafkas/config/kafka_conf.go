// Package config @Author larry
// @Date 2025/1/3 09:15
// @Desc

package config

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"warm-nest/pkg/app/configs"
)

// DefaultCluster 默认集群名，业务不传 cluster 时使用
const DefaultCluster = "default"

// ClusterConfig 单个 Kafka 集群配置
type ClusterConfig struct {
	Urls string `toml:"urls" json:"urls"`
	Auth string `toml:"auth" json:"auth"` // 认证方式：空串=plaintext, "iam"=AWS MSK IAM
	TLS  bool   `toml:"tls" json:"tls"`   // 是否启用 TLS
}

// KafkaConfig 所有集群配置，key 是集群名（例如 "default"、"audit"）
type KafkaConfig map[string]ClusterConfig

// legacyKafkaConfig 旧版单集群配置（kafka.urls）
type legacyKafkaConfig struct {
	Urls string `toml:"urls"`
}

var kafkaConf KafkaConfig
var kafkaConfOnce sync.Once

// kafkaEnvClusterRe 匹配 KAFKA_<CLUSTER>_<FIELD>/kafka.<cluster>.<field> 形式的 env，
// 其中 <FIELD> ∈ {urls, auth, tls}。捕获组 1 = 集群名，捕获组 2 = 字段名。大小写均忽略。
// 限制：集群名不得含 . 或 _，否则无法唯一切分。
var kafkaEnvClusterRe = regexp.MustCompile(`(?i)^kafka[._]([^._]+)[._](urls|auth|tls)$`)

// KafkaConf 获取所有集群配置单例
// 解析策略：
//  1. 新 [kafka.<cluster>] map 与旧 [kafka] urls 并读合并；
//  2. 扫 env 找 toml 没声明的集群（如 KAFKA_FINANCIAL_URLS），直接把 urls/auth/tls
//     写进 cfg —— 绕开 env_loader 对空 map 的早退；
//  3. 若合并后 default.urls 为空则取 legacy.urls 补；两者皆空则 panic。
func KafkaConf() KafkaConfig {
	kafkaConfOnce.Do(func() {
		cfg, _ := configs.Load[KafkaConfig](configs.KafkaConfPath)
		if cfg == nil {
			cfg = KafkaConfig{}
		}
		MergeEnvOnlyClusters(cfg, os.Environ())
		legacy, _ := configs.Load[legacyKafkaConfig](configs.KafkaConfPath)

		def := cfg[DefaultCluster]
		if def.Urls == "" {
			def.Urls = legacy.Urls
		}
		if def.Urls == "" {
			panic("load kafka config failed: neither [kafka.default].urls nor [kafka].urls is set")
		}
		cfg[DefaultCluster] = def
		kafkaConf = cfg
	})
	return kafkaConf
}

// MergeEnvOnlyClusters 扫 envs（一般来源 os.Environ()）把 toml 没声明的集群字段写进 cfg。
// 已声明的集群跳过 —— 那些走 configs.Load 内部的 env_loader 已覆盖完毕。
// 导出给单测注入 envs 切片，避免污染进程环境。
func MergeEnvOnlyClusters(cfg KafkaConfig, envs []string) {
	tomlDeclared := make(map[string]struct{}, len(cfg))
	for name := range cfg {
		tomlDeclared[name] = struct{}{}
	}
	for _, kv := range envs {
		eq := strings.IndexByte(kv, '=')
		if eq <= 0 {
			continue
		}
		m := kafkaEnvClusterRe.FindStringSubmatch(kv[:eq])
		if m == nil {
			continue
		}
		name := strings.ToLower(m[1])
		if _, declared := tomlDeclared[name]; declared {
			continue
		}
		c := cfg[name]
		switch strings.ToLower(m[2]) {
		case "urls":
			c.Urls = kv[eq+1:]
		case "auth":
			c.Auth = kv[eq+1:]
		case "tls":
			c.TLS, _ = strconv.ParseBool(kv[eq+1:])
		}
		cfg[name] = c
	}
}

// GetCluster 查询指定集群配置
func GetCluster(name string) (ClusterConfig, bool) {
	cfg, ok := KafkaConf()[name]
	return cfg, ok
}
