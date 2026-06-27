// Package context @Author larry
// File web_context.go
// @Date 2024/9/26 11:16:00
// @Desc
package context

import (
	"context"

	"warm-nest/pkg/utils/contexts"
	"warm-nest/pkg/utils/traces"
)

const HTTPContextKey = "http-context"

type HTTPContext struct {
	ServiceName string                 // 服务名称
	ReqId       string                 // 请求唯一标识
	CallMsg     map[string]interface{} // 调用信息
	Debug       bool                   // 是否开启debug
	RateInSec   int                    // 每秒请求数
	SignAccount string                 // 签名账户
	IAM         bool                   // 是否对本次请求启用 IAM 签名
	Headers     map[string]string      // 透传给下游服务的自定义请求头
}

func GetHttpContext(ctx context.Context) HTTPContext {
	httpContext, _ := contexts.GetValue[HTTPContext](ctx, HTTPContextKey)
	return httpContext
}

func NewHttpContext(serviceName string, options ...Option) *HTTPContext {
	httpContext := &HTTPContext{
		ServiceName: serviceName,
		ReqId:       traces.TraceId("F"),
		CallMsg:     make(map[string]interface{}),
	}
	for _, option := range options {
		option(httpContext)
	}
	return httpContext
}

func (ctx *HTTPContext) SetContext(options ...Option) *HTTPContext {
	for _, option := range options {
		option(ctx)
	}
	return ctx
}

type Option func(*HTTPContext)

func TraceId(traceId string) Option {
	return func(h *HTTPContext) {
		h.ReqId = traceId
	}
}

func CallMsg(key string, value interface{}) Option {
	return func(h *HTTPContext) {
		if h.CallMsg == nil {
			h.CallMsg = make(map[string]interface{})
		}
		h.CallMsg[key] = value
	}
}

func SignAccount(signAccount string) Option {
	return func(h *HTTPContext) {
		h.SignAccount = signAccount
		h.CallMsg["signAccount"] = signAccount
	}
}

func EnableIAM() Option {
	return func(h *HTTPContext) {
		h.IAM = true
		h.CallMsg["iam"] = true
	}
}

// Header 注入透传给下游服务的自定义请求头，可多次调用追加不同 header
func Header(key, value string) Option {
	return func(h *HTTPContext) {
		if h.Headers == nil {
			h.Headers = make(map[string]string)
		}
		h.Headers[key] = value
	}
}
