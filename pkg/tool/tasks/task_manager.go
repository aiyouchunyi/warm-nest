// Package tasks @Author larry
// @Date 2024/4/19 15:17:00
// @Desc 任务管理器
package tasks

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/app/configs"
	"warm-nest/pkg/tool/tasks/config"
)

type CronManager struct {
	cron  *cron.Cron
	tasks []ITask
}

var enableMu sync.Mutex
var enableCalled bool

func EnableTask(params ...interface{}) (func(), error) {
	enableMu.Lock()
	defer enableMu.Unlock()
	if enableCalled {
		return nil, fmt.Errorf("task manager can only be enabled once")
	}
	enableCalled = true

	if configs.IsLocal() {
		logrus.Warn("task manager is not supported in local environment")
		return nil, nil
	}
	if len(params) != 1 {
		return nil, fmt.Errorf("task manager requires 1 parameters: tasks")
	}
	tasks, ok := params[0].([]ITask)
	if !ok {
		return nil, fmt.Errorf("invalid parameter type for task manager initialization: tasks")
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	taskManager := New(tasks...)
	err := taskManager.Run()
	if err != nil {
		return nil, err
	}

	return taskManager.Stop, nil
}

// New 创建任务管理器
func New(tasks ...ITask) *CronManager {
	logger := cron.VerbosePrintfLogger(logrus.StandardLogger())
	location, _ := time.LoadLocation("Asia/Shanghai")
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(logger),
		cron.WithLocation(location),
		cron.WithChain(
			cron.SkipIfStillRunning(logger), // 任务执行时间过长时，跳过当前任务
			cron.Recover(logger),
		))
	return &CronManager{
		cron:  c,
		tasks: tasks,
	}
}

// Run 启动任务管理器
func (m *CronManager) Run() error {
	logrus.Info("task manager starting...")
	if err := m.Start(); err != nil {
		logrus.WithError(err).Error("task manager start error")
		return err
	}
	logrus.Info("task manager started...")
	return nil
}

// Start 启动任务管理器
func (m *CronManager) Start() error {
	for _, task := range m.tasks {
		taskConfig, ok := config.TaskConf().TaskConfig(task.GetName())
		spec := task.GetSpec()
		if ok {
			if taskConfig.Disabled {
				continue
			}
			if taskConfig.Spec != "" {
				spec = taskConfig.Spec
			}
		}
		_, err := m.cron.AddJob(spec, task)
		if err != nil {
			logrus.WithError(err).Errorf("task manager add %s task error", task.GetName())
			return fmt.Errorf("task manager add %s task error: %+v", task.GetName(), err)
		}
	}
	go func() {
		m.cron.Start()
	}()

	return nil
}

// Stop 停止任务管理器
func (m *CronManager) Stop() {
	stop := m.cron.Stop()
	<-stop.Done()
	logrus.Infof("task manager stoped...")
}
