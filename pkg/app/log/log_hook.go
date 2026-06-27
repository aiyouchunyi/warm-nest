// Package logs @Author larry
// File log_hook.go
// @Date 2024/6/7 13:55:00
// @Desc 日志钩子
package log

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/reflects"
)

var presetMap = map[string]int{
	logrus.FieldKeyTime:        0,
	logrus.FieldKeyLevel:       1,
	logrus.FieldKeyFunc:        2,
	logrus.FieldKeyFile:        3,
	"error":                    4,
	logrus.FieldKeyLogrusError: 5,
	logrus.FieldKeyMsg:         6,
}

type JsonifyHook struct{}

func (h *JsonifyHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *JsonifyHook) Fire(entry *logrus.Entry) error {
	for name, field := range entry.Data {
		if field == nil {
			continue
		}
		if _, exist := presetMap[name]; exist {
			continue
		}
		if name == "traceId" {
			entry.Data[name] = fmt.Sprintf("[%v]", field)
			continue
		}

		if reflects.IsComplex(field) {
			marshal, err := json.Marshal(field)
			if err != nil {
				entry.Data[name] = fmt.Sprintf("%v", field)
			} else {
				entry.Data[name] = fmt.Sprintf("%s", marshal)
			}
		}
	}
	return nil
}
