// Package config @Author larry
// @Date 2025/5/6 15:46
// @Desc

package config

import (
	"strings"
	"sync"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/utils/slices"
)

type DomainConfigs struct {
	Domains []DomainConfig `toml:"domains" json:"domains" gorm:"comment:服务域名配置"`
}

type DomainConfig struct {
	Name   string `toml:"name" json:"name" gorm:"comment:渠道或者服务名"`
	Domain string `toml:"domain" json:"domain" gorm:"comment:域名"`
}

var domainConf DomainConfigs
var domainConfOnce sync.Once

// DomainConf 获取服务域名配置单例
func DomainConf() DomainConfigs {
	domainConfOnce.Do(func() {
		domainConf = configs.MustLoad[DomainConfigs](configs.HttpConfPath)
	})
	return domainConf
}

func Init(conf DomainConfigs) {
	domainConfOnce.Do(func() {
		domainConf = conf
	})
}

// GetDomain 获取域名配置
func (conf DomainConfigs) GetDomain(name string) (DomainConfig, bool) {
	return slices.FindOne(conf.Domains, func(domain DomainConfig) bool {
		return strings.EqualFold(domain.Name, name)
	})
}
