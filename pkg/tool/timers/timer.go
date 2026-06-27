// Package timers @Author larry
// @Date 2025/4/8 14:48
// @Desc

package timers

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/times"
	"warm-nest/pkg/utils/traces"
)

type Executor func() error

func Timer(name string, timeInSec int64, f Executor) {
	go func() {
		logrus.Infof("%s timer started...", name)
		ticker := time.NewTicker(time.Second * time.Duration(timeInSec))
		defer ticker.Stop()
		for range ticker.C {
			traceId := traces.TraceId("TI")
			start := times.UnixMilli()
			logrus.Infof("[%s] %s timer running...", traceId, name)
			func() {
				defer func() {
					if r := recover(); r != nil {
						logrus.Error(fmt.Sprintf("[%s] %s timer panic! delay:%s panic:%v", traceId, name, times.GapMS(start), r))
					}
				}()
				if err := f(); err != nil {
					logrus.Error(fmt.Sprintf("[%s] %s timer failed! delay:%s err:%v", traceId, name, times.GapMS(start), err))
				}
			}()
			logrus.Infof("[%s] %s timer finished... delay:%s", traceId, name, times.GapMS(start))
		}
	}()
}
