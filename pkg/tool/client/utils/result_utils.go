// Package utils @Author larry
// File result_utils.go
// @Date 2024/9/26 15:24:00
// @Desc
package utils

import (
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/client/code"
	"warm-nest/pkg/tool/client/context"
	"warm-nest/pkg/tool/client/model"
)

// NilResult 解析结果
func NilResult(res *resty.Response, err error) {
	_ = NoResult(res, err)
}

// NoResult 解析结果
func NoResult(res *resty.Response, err error) error {
	if err == nil {
		return err
	}
	BusinessError(res, err)
	return err
}

// Result 解析结果
func Result[R model.IRespData[D], D any](resp *resty.Response, err error) (D, error) {
	var d D
	if err != nil {
		return d, NoResult(resp, err)
	}
	result, ok := resp.Result().(*R)
	if !ok {
		logrus.Errorf("Result type err! %v", resp.Result())
		return d, errors.NewWithArgs(code.HttpResponseParseErr, resp.Result())
	}
	if result == nil {
		return d, nil
	}
	return (*result).GetData(), nil
}

func ResultArray[R model.IRespData[[]D], D any](resp *resty.Response, err error) ([]D, error) {
	a, err := ResultPtr[R, []D](resp, err)
	if a != nil {
		return *a, err
	}
	var d []D
	return d, err
}

func ResultArray0[R model.IRespData[[]D], D any](resp *resty.Response, err error) (D, error) {
	a, err := ResultArray[R, D](resp, err)
	if len(a) != 0 {
		return a[0], err
	}
	var d D
	return d, err
}

// ResultPtr 解析结果
func ResultPtr[R model.IRespData[D], D any](resp *resty.Response, err error) (*D, error) {
	result, err := Result[R, D](resp, err)
	return &result, err
}

// ResultPtrV2 解析结果
func ResultPtrV2[R any](resp *resty.Response, err error) (*R, error) {
	if err != nil {
		return nil, NoResult(resp, err)
	}
	result, ok := resp.Result().(*R)
	if !ok {
		logrus.Errorf("Result type err! %v", resp.Result())
		return nil, errors.NewWithArgs(code.HttpResponseParseErr, resp.Result())
	}
	return result, nil
}

// ResultV2 解析结果
func ResultV2[R any](resp *resty.Response, err error) (R, error) {
	ptr, err := ResultPtrV2[R](resp, err)
	if ptr != nil {
		return *ptr, err
	}
	var d R
	return d, err
}

// ResultArrayV2 解析结果
func ResultArrayV2[R any](resp *resty.Response, err error) ([]R, error) {
	ptr, err := ResultPtrV2[[]R](resp, err)
	if ptr != nil {
		return *ptr, err
	}
	var d []R
	return d, err
}

// ResultArray0V2 解析结果
func ResultArray0V2[R any](resp *resty.Response, err error) (R, error) {
	ptr, err := ResultArrayV2[R](resp, err)
	if len(ptr) != 0 {
		return ptr[0], err
	}
	var d R
	return d, err
}

// BusinessError 记录结果
func BusinessError(res *resty.Response, err error) {
	if res == nil {
		logrus.WithError(err).Errorf("request failed!")
		return
	}
	httpContext := context.GetHttpContext(res.Request.Context())
	log := logrus.WithError(err)
	if len(httpContext.CallMsg) != 0 {
		log = log.WithFields(httpContext.CallMsg)
	}
	log.WithField("requestId", httpContext.ReqId).
		WithFields(logrus.Fields{"url": res.Request.URL}).
		Errorf("[Feign-%s-Err] business err!", httpContext.ServiceName)
}
