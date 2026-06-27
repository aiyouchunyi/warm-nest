// Package config @Author larry
// @Date 2025/5/6 11:28
// @Desc

package config

import (
	"net/http"
	"strings"
	"sync"

	"github.com/creasty/defaults"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/tool/client/sign"
	"warm-nest/pkg/utils/slices"
)

type ClientConfigs struct {
	Clients []ClientConfig `toml:"clients" json:"clients" gorm:"comment:HTTP客户端配置"`
}

var clientConfig ClientConfigs
var clientConfOnce sync.Once

// ClientConf 获取HTTP客户端配置单例
func ClientConf() ClientConfigs {
	clientConfOnce.Do(func() {
		clientConfig = configs.MustLoad[ClientConfigs](configs.HttpConfPath)
	})
	return clientConfig
}

// GetClientConfig 获取HTTP客户端配置
func (conf ClientConfigs) GetClientConfig(serviceName string) ClientConfig {
	config := slices.FindOrGet(conf.Clients, func(config ClientConfig) bool {
		return strings.EqualFold(config.ServiceName, serviceName)
	}, func() ClientConfig {
		return ClientConfig{ServiceName: serviceName}
	})
	_ = defaults.Set(&config)
	return config
}

type ClientConfig struct {
	ServiceName string            `toml:"serviceName" json:"serviceName" gorm:"comment:服务名称"`
	Timeout     int               `toml:"timeout" json:"timeout" gorm:"comment:超时时间s" default:"256"`
	RateLimit   int               `toml:"rateLimit" json:"rateLimit" gorm:"comment:限流次数" default:"256"`
	BodySize    int               `toml:"bodySize" json:"bodySize" gorm:"comment:body日志大小" default:"1024"`
	Signature   sign.DoSignature  `gorm:"comment:签名函数"`
	Transport   http.RoundTripper `gorm:"-"` // 自定义底层 Transport（用于 IAM 签名等场景）
}

func NewConfig(serviceName string, options ...Option) ClientConfig {
	config := ClientConf().GetClientConfig(serviceName)
	for _, option := range options {
		option(&config)
	}
	return config
}

type Option func(*ClientConfig)

func WithSignature(signature sign.DoSignature) Option {
	return func(c *ClientConfig) {
		c.Signature = signature
	}
}

func WithTimeout(timeout int) Option {
	return func(c *ClientConfig) {
		c.Timeout = timeout
	}
}

func WithRateLimit(rateLimit int) Option {
	return func(c *ClientConfig) {
		c.RateLimit = rateLimit
	}
}

func WithBodySize(bodySize int) Option {
	return func(c *ClientConfig) {
		c.BodySize = bodySize
	}
}

func WithTransport(transport http.RoundTripper) Option {
	return func(c *ClientConfig) {
		c.Transport = transport
	}
}
