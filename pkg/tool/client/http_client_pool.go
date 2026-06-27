// Package client @Author larry
// File binance_client.go
// @Date 2024/4/26 14:08:00
// @Desc Http多渠道客户端
package client

import (
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/client/config"
	"warm-nest/pkg/tool/client/model"
	"warm-nest/pkg/tool/client/sign"
	"warm-nest/pkg/tool/client/utils"
	"warm-nest/pkg/tool/concurrent"
	"warm-nest/pkg/utils/reflects"
)

var httpClientPool = concurrent.NewCMap[*HTTPClient]()

func RegisterDefault(serviceName string, options ...config.Option) {
	Register[model.RespCode](serviceName, options...)
}

func Register[e model.IRespCode](serviceName string, options ...config.Option) {
	httpClientPool.SetIfAbsent(serviceName, func() *HTTPClient {
		httpConfig := config.NewConfig(serviceName, options...)
		httpClient := New(httpConfig)
		httpClient.SetError(reflects.New[e]())
		httpClient.OnAfterResponse(utils.ParseErr[e])
		if httpConfig.Signature != nil {
			httpClient.SetPreRequestHook(sign.DoSign(httpConfig.Signature))
		}
		return httpClient
	})
}

// GetHttpClient 获取客户端
// service 服务名
// return 客户端
func GetHttpClient(serviceName string) *HTTPClient {
	client, ok := httpClientPool.Load(serviceName)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"serviceName": serviceName,
		}).Error("client unregistered!")
		panic("client unregistered! serviceName=" + serviceName)
	}
	return client
}
