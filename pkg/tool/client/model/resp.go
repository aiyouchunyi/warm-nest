// Package model @Author larry
// File binance_resp.go
// @Date 2024/9/19 14:08:00
// @Desc
package model

import (
	"fmt"
)

type IRespCode interface {
	Success() bool      // 是否成功
	GetCode() string    // 获取错误码
	GetMessage() string // 获取错误信息
}

type IRespData[V any] interface {
	GetData() V
}

type Resp[V any] struct {
	RespCode
	Data V `json:"data"`
}

func (r Resp[V]) GetData() V {
	return r.Data
}

type RespCode struct {
	Result    bool   `json:"result"`
	Mcode     string `json:"mcode"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func (r RespCode) Success() bool {
	return r.Result
}

func (r RespCode) GetCode() string {
	return r.Mcode
}

func (r RespCode) GetMessage() string {
	return r.Message
}

type RespV2[V any] struct {
	RespCodeV2
	Data V `json:"data"`
}

func (r RespV2[V]) GetData() V {
	return r.Data
}

type RespCodeV2 struct {
	Result    bool   `json:"result"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func (r RespCodeV2) Success() bool {
	return r.Result
}

func (r RespCodeV2) GetCode() string {
	return r.Code
}

func (r RespCodeV2) GetMessage() string {
	return r.Message
}

type RespV3[V any] struct {
	RespCodeV3
	Data V `json:"data"`
}

func (r RespV3[V]) GetData() V {
	return r.Data
}

type RespCodeV3 struct {
	Result    bool   `json:"result"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func (r RespCodeV3) Success() bool {
	return r.Result || r.Code == 0
}

func (r RespCodeV3) GetCode() string {
	return fmt.Sprintf("%d", r.Code)
}

func (r RespCodeV3) GetMessage() string {
	return r.Message
}
