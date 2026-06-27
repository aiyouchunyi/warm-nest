// Package tasks @Author Larry
// @Date 2024/10/22 09:14
// @Desc

package tasks

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/utils/times"
	"warm-nest/pkg/utils/traces"
)

type Executor func(args ...any) error

type ITask interface {
	cron.Job
	GetName() string // 任务名称
	GetSpec() string // 任务规则
}

type Task struct {
	spec     string
	name     string
	executor Executor
	args     []any
}

// NewTask 创建任务
func NewTask(spec string, name string, executor Executor, args ...any) *Task {
	return &Task{
		spec:     spec,
		name:     name,
		executor: executor,
		args:     args,
	}
}

// GetName 获取定时器名称
func (t *Task) GetName() string {
	return t.name
}

// GetSpec 获取定时器规则
func (t *Task) GetSpec() string {
	return t.spec
}

// Run 启动任务状态监控定时器
func (t *Task) Run() {
	traceId := traces.TraceId("TA")
	start := times.UnixMilli()
	logrus.Infof("[%s] %s task running...", traceId, t.GetName())
	defer func() {
		if r := recover(); r != nil {
			logrus.Error(fmt.Sprintf("[%s] %s task panic! delay:%s panic:%v", traceId, t.GetName(), times.GapMS(start), r))
		}
	}()
	if err := t.executor(t.args...); err != nil {
		logrus.Error(fmt.Sprintf("[%s] %s task failed! delay:%s err:%v", traceId, t.GetName(), times.GapMS(start), err))
		return
	}
	logrus.Infof("[%s] %s task finished... delay:%s", traceId, t.GetName(), times.GapMS(start))
}
