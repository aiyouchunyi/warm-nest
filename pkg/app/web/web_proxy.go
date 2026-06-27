// Package web @Author larry
// File web_proxy.go
// @Date 2024/8/13 11:02:00
// @Desc 代理
package web

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/app/web/resp"
	"warm-nest/pkg/app/web/widgets"
)

type WebProxy struct {
	domain     string   // 默认代理域名
	prodDomain string   // 生产域名
	redirects  []string // 重定向移除路径
}

func NewProxy(domain string, redirects ...string) *WebProxy {
	return &WebProxy{
		domain:    domain,
		redirects: redirects,
	}
}

func NewProxyMulDomain(domain string, prodDomain string, redirects ...string) *WebProxy {
	return &WebProxy{
		domain:     domain,
		prodDomain: prodDomain,
		redirects:  redirects,
	}
}

func (p *WebProxy) RedirectProxy(handles ...widgets.WidgetFunc) func(c *gin.Context) {
	return p.Proxy("", handles...)
}

func (p *WebProxy) Proxy(proxyPath string, handles ...widgets.WidgetFunc) func(c *gin.Context) {
	return func(c *gin.Context) {
		for _, handle := range handles {
			if err := handle(c); err != nil {
				resp.Fail(c, err)
				return
			}
		}
		p.proxy(proxyPath, c)
	}
}

// proxy 实现接口转发
func (p *WebProxy) proxy(proxyPath string, c *gin.Context) {
	targetPath := proxyPath
	if targetPath == "" {
		targetPath = c.Request.URL.Path
		for _, redirect := range p.redirects {
			targetPath = strings.ReplaceAll(targetPath, redirect, "")
		}
	}

	domain := p.domain
	if configs.IsProd() && p.prodDomain != "" {
		domain = p.prodDomain
	}

	remote, _ := url.Parse(domain)
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.RawQuery = c.Request.URL.RawQuery
		req.URL.Host = remote.Host
		req.URL.Path = targetPath
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}
