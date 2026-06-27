// Package client @Author larry
// File binance_client.go
// @Date 2024/4/26 14:08:00
// @Desc Http客户端
package client

import (
	errors2 "errors"
	"io"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/juju/ratelimit"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/client/code"
	"warm-nest/pkg/tool/client/config"
	"warm-nest/pkg/tool/client/context"
	"warm-nest/pkg/tool/client/model"
	"warm-nest/pkg/tool/client/utils"
	"warm-nest/pkg/tool/ratelimits"
	"warm-nest/pkg/utils/transforms"
)

type RateBucket struct {
	bucket *ratelimit.Bucket
}

func (b RateBucket) Allow() bool {
	return b.bucket.WaitMaxDuration(1, time.Duration(10)*time.Second)
}

type HTTPClient struct {
	httpConfig config.ClientConfig
	*resty.Client
}

// New 创建http客户端
func New(httpConfig config.ClientConfig) *HTTPClient {
	restyClient := resty.New()
	if httpConfig.Transport != nil {
		restyClient.SetTransport(httpConfig.Transport)
	}
	return &HTTPClient{
		httpConfig: httpConfig,
		Client: restyClient.
			SetBaseURL(Resolve(httpConfig.ServiceName)).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				if err != nil {
					return strings.Contains(err.Error(), "unexpected EOF") ||
						errors2.Is(err, io.ErrUnexpectedEOF) ||
						errors2.Is(err, syscall.ECONNRESET)
				}
				return r.StatusCode() >= 500
			}).
			AddRetryHook(func(resp *resty.Response, err error) {
				httpContext := context.GetHttpContext(resp.Request.Context())
				retry := 0
				url := ""
				method := ""
				if resp != nil && resp.Request != nil {
					retry = resp.Request.Attempt // 第几次尝试（从 1 开始）
					method = resp.Request.Method
					url = resp.Request.URL
				}
				log := logrus.WithField("requestId", httpContext.ReqId).WithFields(httpContext.CallMsg).WithFields(logrus.Fields{
					"method": method,
					"retry":  retry,
					"url":    url,
				})
				if err != nil {
					log.WithError(err).Errorf("[Fegin-%s-Retry]", httpContext.ServiceName)
				} else {
					log.Infof("[Fegin-%s-Retry]", httpContext.ServiceName)
				}
			}).
			SetLogger(logrus.StandardLogger()).
			SetTimeout(time.Duration(httpConfig.Timeout) * time.Second).
			SetRateLimiter(RateBucket{
				bucket: ratelimit.NewBucketWithQuantum(time.Second, int64(httpConfig.RateLimit)+1, int64(httpConfig.RateLimit)),
			}).
			SetDebugBodyLimit(int64(httpConfig.BodySize)).
			OnBeforeRequest(func(_ *resty.Client, request *resty.Request) error {
				httpContext := context.GetHttpContext(request.Context())
				if len(httpContext.Headers) > 0 {
					request.SetHeaders(httpContext.Headers)
				}
				return nil
			}).
			OnBeforeRequest(func(c *resty.Client, request *resty.Request) error {
				httpContext := context.GetHttpContext(request.Context())
				if httpContext.RateInSec <= 0 {
					return nil
				}
				path := utils.ParsePath(c, request)
				ratelimits.Wait(path, httpContext.RateInSec)
				return nil
			}).
			OnBeforeRequest(func(c *resty.Client, request *resty.Request) error {
				httpContext := context.GetHttpContext(request.Context())
				if httpContext.Debug {
					return nil
				}
				log := logrus.WithField("requestId", httpContext.ReqId).WithFields(httpContext.CallMsg).
					WithFields(logrus.Fields{
						"method": request.Method,
						"url":    utils.ParseUrl(c, request),
					})
				if request.Body != nil {
					bodyStr := transforms.Marshal(request.Body)
					if len(bodyStr) > httpConfig.BodySize {
						bodyStr = "req body size too large!"
					}
					log = log.WithFields(logrus.Fields{
						"body": bodyStr,
					})
				}
				log.Infof("[Fegin-%s-Call]", httpContext.ServiceName)
				return nil
			}).
			OnAfterResponse(func(c *resty.Client, response *resty.Response) error {
				httpContext := context.GetHttpContext(response.Request.Context())
				if httpContext.Debug {
					return nil
				}

				log := logrus.WithField("requestId", httpContext.ReqId).WithFields(httpContext.CallMsg).
					WithFields(logrus.Fields{
						"status":       response.StatusCode(),
						"Content-Type": response.Header().Get("Content-Type"),
						"url":          response.Request.URL,
						"delay":        response.Time(),
					})

				bodyStr := response.String()
				if len(bodyStr) > httpConfig.BodySize {
					bodyStr = "resp body size too large!"
				}
				if len(bodyStr) != 0 {
					log = log.WithFields(logrus.Fields{
						"body": bodyStr,
					})
				}

				if response.IsSuccess() {
					log.Infof("[Fegin-%s-Succ]", httpContext.ServiceName)
					return nil
				}
				log.Errorf("[Fegin-%s-Fail]", httpContext.ServiceName)
				return nil
			}).
			OnAfterResponse(func(_ *resty.Client, response *resty.Response) error {
				if response.IsSuccess() {
					return nil
				}
				if response.Error() != nil {
					respCode, ok := response.Error().(model.IRespCode)
					if ok && respCode.GetCode() != "" {
						return nil
					}
				}
				// todo: add 明确错误码
				if response.StatusCode() == http.StatusTooManyRequests {
					return errors.NewWithArgs(code.TooManyRequests, response.Request.URL)
				}
				if response.IsError() {
					return errors.NewWithArgs(code.HttpResponseFailed, response.StatusCode(), response.String())
				}
				return errors.NewWithArgs(code.HttpResponseUnknown, response.StatusCode(), response.String())
			}),
	}
}
