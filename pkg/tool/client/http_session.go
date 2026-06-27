// Package client @Author larry
// @Date 2025/2/28 16:02
// @Desc

package client

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/client/context"
)

type HttpSession struct {
	*context.HTTPContext
}

func NewSession(serviceName string, options ...context.Option) *HttpSession {
	return &HttpSession{
		context.NewHttpContext(serviceName, options...),
	}
}

// Request 请求
func (session *HttpSession) Request(options ...context.Option) *HTTPRequest {
	return NewHTTPRequest(session.SetContext(options...), GetHttpClient(session.ServiceName))
}

// SignRequest 请求
func (session *HttpSession) SignRequest(signAccount string, options ...context.Option) *HTTPRequest {
	return NewHTTPRequest(session.SetContext(append(options, context.SignAccount(signAccount))...), GetHttpClient(session.ServiceName))
}

// IAMRequest 发起启用 IAM 签名的请求
func (session *HttpSession) IAMRequest(options ...context.Option) *HTTPRequest {
	return NewHTTPRequest(session.SetContext(append(options, context.EnableIAM())...), GetHttpClient(session.ServiceName))
}

func (session *HttpSession) Log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"service": session.ServiceName,
		"reqId":   session.ReqId,
	}).WithFields(session.CallMsg)
}
