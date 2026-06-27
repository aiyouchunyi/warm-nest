// Package utils @Author larry
// File url_utils.go
// @Date 2024/9/19 14:37:00
// @Desc
package utils

import (
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

// ParseUri 解析请求uri
func ParseUri(request *http.Request) string {
	if query := request.URL.RawQuery; query == "" {
		return request.URL.Path
	}
	return request.URL.Path + "?" + request.URL.Query().Encode()
}

func ParseUrl(c *resty.Client, request *resty.Request) string {
	url := ParsePath(c, request)
	if query := request.QueryParam.Encode(); query != "" {
		url += "?" + query
	}
	return url
}

func ParsePath(c *resty.Client, request *resty.Request) string {
	domainPath := request.URL
	if !strings.Contains(request.URL, "http") {
		domainPath = c.BaseURL + request.URL
	}
	return domainPath
}
