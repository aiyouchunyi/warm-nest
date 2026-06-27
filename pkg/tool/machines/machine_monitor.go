// Package machines @Author larry
// @Date 2025/2/8 13:45
// @Desc

package machines

import (
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/numbers"
)

// AsyncMonitor 监控
func (m *Machine[T]) AsyncMonitor(_ ...any) error {
	// 1.查询监控任务
	tasks, err := m.MonitorTasks()
	if err != nil {
		return err
	}

	// 2.执行监控任务
	m.Submits(tasks)
	return nil
}

// Monitor 监控
func (m *Machine[T]) Monitor(args ...any) error {
	sleepTime := 0
	if len(args) > 0 {
		sleepTime = numbers.MustInt(numbers.NewT(args[0]))
	}

	// 1.查询监控任务
	tasks, err := m.MonitorTasks()
	if err != nil {
		return err
	}

	// 2.执行监控任务
	for index := range tasks {
		task := tasks[index]
		executeInfo, err2 := m.Execute(task)
		if err2 != nil {
			logrus.WithFields(logrus.Fields{
				"executeInfo": executeInfo,
			}).WithError(err2).Errorf("Machine execute task failed!")
		}
		if sleepTime > 0 {
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}
	return nil
}
