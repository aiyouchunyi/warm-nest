// Package interceptor @Author larry
// @Date 2024/4/15 15:20:00
// @Desc web请求日志

package interceptor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/web/context"
	"warm-nest/pkg/app/web/utils"
	"warm-nest/pkg/utils/slices"
	"warm-nest/pkg/utils/times"
)

var FilterUrls = []string{
	"/ping",
}

const maxResponseBodySize = 1024 // 响应 body 最大打印大小（字节）

// responseWriter 自定义响应写入器，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写 Write 方法，捕获响应内容（w: 写入数据, 返回: 写入字节数和错误）
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// addResponseBodyField 添加响应 body 日志字段（body: 响应内容缓冲区, fields: 日志字段集合）
func addResponseBodyField(body *bytes.Buffer, fields logrus.Fields) {
	if body.Len() == 0 {
		return
	}
	bodySize := body.Len()
	if bodySize <= maxResponseBodySize {
		fields["body"] = body.String()
	} else {
		fields["body"] = fmt.Sprintf("%d bytes", bodySize)
	}
}

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		url := c.Request.URL.String()
		if slices.Contain(FilterUrls, url) {
			return
		}

		start := times.UnixMilli()
		webContext := context.GetContext(c)

		log := logrus.WithField("reqId", webContext.ReqId).WithFields(logrus.Fields{
			"reqUser": webContext.ReqUser,
			"url":     utils.UrlInfo(c),
		})
		log.WithFields(logrus.Fields{
			"body": Body(c),
		}).Infof(" =====>")

		// 使用自定义 responseWriter 捕获响应内容
		respWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		c.Writer = respWriter

		// Process request
		c.Next()

		// 记录响应日志
		fields := logrus.Fields{
			"status": c.Writer.Status(),
			"delay":  times.GapMS(start),
		}
		addResponseBodyField(respWriter.body, fields)
		log.WithFields(fields).Infof(" <=====")
	}
}

func Body(c *gin.Context) string {
	if c.Request.ContentLength == 0 {
		return "empty"
	}
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "empty"
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var compacted bytes.Buffer
	if err2 := json.Compact(&compacted, bodyBytes); err2 != nil {
		return "empty"
	}

	bodyStr := compacted.String()
	if len(bodyStr) > maxResponseBodySize {
		bodyStr = "req body size too large!"
	}
	return bodyStr
}
