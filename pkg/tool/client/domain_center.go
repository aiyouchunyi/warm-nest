// Package client @Author larry
// File domain_center.go
// @Date 2024/6/12 17:13:00
// @Desc 服务域名配置中心
package client

import (
	strings2 "strings"
	"sync"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/tool/client/config"
	"warm-nest/pkg/utils/slices"
)

type DomainCenter struct {
	defaultDomains []config.DomainConfig
	prodDomains    []config.DomainConfig
}

func GetDomainCenter() *DomainCenter {
	domainCenterOnce.Do(func() {
		domainCenter = &DomainCenter{
			defaultDomains: make([]config.DomainConfig, 0),
			prodDomains:    make([]config.DomainConfig, 0),
		}
	})
	return domainCenter
}

var domainCenter *DomainCenter
var domainCenterOnce sync.Once

// RegisterDefaultDomain 注册默认域名
func RegisterDefaultDomain(domainConfig config.DomainConfig) {
	GetDomainCenter().defaultDomains = append(GetDomainCenter().defaultDomains, domainConfig)
}

// RegisterProdDomain 添加生产环境域名
func RegisterProdDomain(domainConfig config.DomainConfig) {
	GetDomainCenter().prodDomains = append(GetDomainCenter().prodDomains, domainConfig)
}

// Resolve 解析域名
func Resolve(name string) string {
	if configDomain, exist := config.DomainConf().GetDomain(name); exist {
		return configDomain.Domain
	}
	if configs.IsProd() {
		if configDomain, exist := slices.FindOne(GetDomainCenter().prodDomains, func(v config.DomainConfig) bool {
			return strings2.EqualFold(v.Name, name)
		}); exist {
			return configDomain.Domain
		}
	}
	if configDomain, exist := slices.FindOne(GetDomainCenter().defaultDomains, func(v config.DomainConfig) bool {
		return strings2.EqualFold(v.Name, name)
	}); exist {
		return configDomain.Domain
	}
	return "http://internalgw:8080"
}
