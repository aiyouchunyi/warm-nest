// Package logs @Author larry
// @Date 2024/4/15 10:53:00
// @Desc 日志处理
package log

import (
	"fmt"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/log/config"
	"warm-nest/pkg/utils/strings"
)

func Init() {
	conf := config.LoggerConf()
	switch conf.Format {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: conf.TimeFormat,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				return fmt.Sprintf("%s:%s:%d", path.Base(f.File), path.Base(f.Function), f.Line), ""
			},
		})
	case "text":
		logrus.SetFormatter(
			&logrus.TextFormatter{
				TimestampFormat: conf.TimeFormat,
				FullTimestamp:   true,
				DisableQuote:    true,
				ForceColors:     true,
				CallerPrettyfier: func(f *runtime.Frame) (string, string) {
					return fmt.Sprintf("[%s:%s:%d]", path.Base(f.File), strings.GetLastDot(f.Function), f.Line), ""
				},
			})
		logrus.AddHook(&JsonifyHook{})
	default:
		panic("logger format un-support: " + conf.Format)
	}

	if conf.Level != "" {
		logLevel, err := logrus.ParseLevel(conf.Level)
		if err != nil {
			logrus.Warnf("cannot parse logger level %s, using 'bo' as default", conf.Level)
			logLevel = logrus.InfoLevel
		}
		logrus.SetLevel(logLevel)
	}
	logrus.SetReportCaller(true)
	logrus.Infof("init logs succ... level=%s format=%s", conf.Level, conf.Format)
}
