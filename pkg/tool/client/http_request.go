// Package client @Author larry
// File http_request.go
// @Date 2024/5/16 09:11:00
// @Desc Http请求
package client

import (
	"github.com/go-resty/resty/v2"

	"warm-nest/pkg/tool/client/context"
	"warm-nest/pkg/tool/client/utils"
	"warm-nest/pkg/utils/contexts"
)

type HTTPRequest struct {
	*resty.Request
}

func NewHTTPRequest(ctx *context.HTTPContext, client *HTTPClient) *HTTPRequest {
	r := &HTTPRequest{
		Request: client.R(),
	}
	r.SetContext(contexts.SetValue(r.Context(), context.HTTPContextKey, ctx))
	return r
}

// SetQuery 设置查询参数
func (r *HTTPRequest) SetQuery(req interface{}) *HTTPRequest {
	r.SetQueryParams(utils.ToQuery(req))
	return r
}

// Debug 设置Debug日志
func (r *HTTPRequest) Debug() *HTTPRequest {
	r.SetContext(contexts.WithValue[context.HTTPContext](r.Context(), context.HTTPContextKey, func(httpContext *context.HTTPContext) {
		httpContext.Debug = true
	}))
	r.Request.SetDebug(true)
	return r
}

func (r *HTTPRequest) Limit(rateInSec int) *HTTPRequest {
	r.SetContext(contexts.WithValue[context.HTTPContext](r.Context(), context.HTTPContextKey, func(httpContext *context.HTTPContext) {
		httpContext.RateInSec = rateInSec
	}))
	return r
}

// CallMsg 设置调用信息
func (r *HTTPRequest) CallMsg(key string, value interface{}) *HTTPRequest {
	r.SetContext(contexts.WithValue[context.HTTPContext](r.Context(), context.HTTPContextKey, func(httpContext *context.HTTPContext) {
		if len(httpContext.CallMsg) == 0 {
			httpContext.CallMsg = make(map[string]interface{})
		}
		httpContext.CallMsg[key] = value
	}))
	return r
}
