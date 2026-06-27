// Package api @Author larry
// @Date 2026/06/21
// @Desc 收货地址簿接口聚合（struct + singleton）

package api

import "sync"

// ApiAddress 收货地址簿接口
type ApiAddress struct{}

var apiAddress *ApiAddress
var apiAddressOnce sync.Once

// GetApiAddress 获取收货地址簿接口单例
func GetApiAddress() *ApiAddress {
	apiAddressOnce.Do(func() {
		apiAddress = &ApiAddress{}
	})
	return apiAddress
}
