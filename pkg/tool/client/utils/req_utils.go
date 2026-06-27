// Package utils @Author larry
// @Date 2025/11/24 23:23
// @Desc

package utils

import (
	"bytes"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

func GetPayload(request *http.Request) string {
	var payload string
	if request.Method == http.MethodGet {
		payload = request.URL.RawQuery
	} else if request.Method == http.MethodPost {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			logrus.WithError(err).Errorf("read request body failed!")
			return ""
		}
		payload = string(bodyBytes)
		request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset the body
	}
	return payload
}
