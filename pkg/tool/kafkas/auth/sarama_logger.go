// Package auth @Author larry
// @Date 2026-05-19
// @Desc 把 sarama 内部日志重定向到 logrus，便于排查 fetch / SASL / broker 异常

package auth

import (
	"strings"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

// init 在包加载时把 sarama.Logger 替换为 logrus 适配器。
// sarama 默认 Logger 是 io.Discard 包装，关键诊断信息（fetch session 失败、broker disconnect、
// SASL 重协商）都丢弃了；接管后能在日志看到完整链路。
// 放 auth 包是因为它被 consumer + producer 都 import，覆盖所有 sarama 路径。
func init() {
	sarama.Logger = saramaLogger{}
}

// saramaLogger 把 sarama 的 stdlib 风格 Print/Printf/Println 适配到 logrus，
// 统一打 source=sarama 字段方便检索。EOF / disconnect 等都按 Info 级别输出，
// 真正的 ERROR 由业务侧 ConsumerError 通道处理。
type saramaLogger struct{}

func (saramaLogger) Print(v ...interface{}) {
	logrus.WithField("source", "sarama").Info(v...)
}

func (saramaLogger) Printf(format string, v ...interface{}) {
	logrus.WithField("source", "sarama").Infof(strings.TrimRight(format, "\n"), v...)
}

func (saramaLogger) Println(v ...interface{}) {
	logrus.WithField("source", "sarama").Info(v...)
}
