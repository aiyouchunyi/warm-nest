// Package sign @Author larry
// @Date 2025/11/14 16:28
// @Desc

package sign

import (
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/client/code"
	"warm-nest/pkg/tool/client/context"
)

type DoSignature func(signAccount string, request *http.Request) error

// DoSign 签名
func DoSign(signature DoSignature) func(_ *resty.Client, request *http.Request) error {
	return func(_ *resty.Client, request *http.Request) error {
		httpContext := context.GetHttpContext(request.Context())
		if httpContext.SignAccount == "" {
			return nil
		}
		err := signature(httpContext.SignAccount, request)
		if err != nil {
			if strings.Contains(err.Error(), "sign account absent") {
				return errors.NewWithArgs(code.SignAccountAbsent, httpContext.SignAccount)
			}
			return err
		}
		return nil
	}
}
