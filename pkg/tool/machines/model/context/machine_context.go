// Package context provides the execution context for state machines.
//
// MachineContext encapsulates:
// - Task identification (ApproveId, TaskName, TaskField)
// - Current state (Status, Lock, StatusInfos)
// - Runtime variables
// - Logging foundation (Log method with traceId)
//
// @Author larry
// File machine_context.go
// @Date 2024/5/7 14:49:00
package context

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"warm-nest/pkg/tool/machines/model/config"
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/task"
	"warm-nest/pkg/tool/machines/model/task/info"
	"warm-nest/pkg/tool/machines/model/variable"
	"warm-nest/pkg/utils/numbers"
	"warm-nest/pkg/utils/strings"
	"warm-nest/pkg/utils/traces"
)

type MachineContext struct {
	traceId           string
	Change            bool                 // 状态是否变更
	MachineStatus                          // 状态机状态配置
	task.TaskIdentify                      // 状态机唯一标识
	*task.MachineTask                      // 状态信息
	Config            config.MachineConfig // 状态机配置
	Task              task.IMachineTask    // 执行任务
}

// NewContext 初始化状态机上下文
func NewContext[T task.IMachineTask](task T, machineStatus MachineStatus, cfg config.MachineConfig, options ...variable.Option) *MachineContext {
	return NewTraceContext(traces.TraceId18("M-"+task.GetIdentify().TaskId), task, machineStatus, cfg, options...)
}

// NewTraceContext 初始化状态机上下文
func NewTraceContext[T task.IMachineTask](traceId string, task T, machineStatus MachineStatus, cfg config.MachineConfig, options ...variable.Option) *MachineContext {
	return &MachineContext{
		traceId:       traceId,
		TaskIdentify:  task.GetIdentify(),
		MachineTask:   task.GetTask(options...),
		MachineStatus: machineStatus,
		Config:        cfg,
		Task:          task,
	}
}

// ToTask 获取信息
func ToTask[O any](c *MachineContext) O {
	return c.Task.(O)
}

// State 获取当前任务
func (ctx *MachineContext) State() string {
	return fmt.Sprintf("%s[%s-%s]", ctx.TaskName, ctx.TaskId, ctx.Status)
}

// Log 返回带任务上下文的日志器（包含traceId和任务状态）
// 这是所有任务相关日志的基础方法
func (ctx *MachineContext) Log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"traceId": ctx.traceId,
		"task":    ctx.State(),
	})
}

// IsHaltStatus 是否是停止状态
func (ctx *MachineContext) IsHaltStatus() bool {
	return ctx.MachineStatus.IsHaltStatus(ctx.Status)
}

// IsFinalStatus 是否是最终状态
func (ctx *MachineContext) IsFinalStatus() bool {
	return ctx.MachineStatus.IsFinalStatus(ctx.Status)
}

func (ctx *MachineContext) IsManualStatus() bool {
	return ctx.MachineStatus.IsManualStatus(ctx.Status)
}

// IsSuccessStatus 是否是成功状态
func (ctx *MachineContext) IsSuccessStatus() bool {
	return ctx.MachineStatus.IsSuccessStatus(ctx.Status)
}

// IsFailStatus 是否是失败状态
func (ctx *MachineContext) IsFailStatus() bool {
	return ctx.MachineStatus.IsFailStatus(ctx.Status)
}

// IsAutoStatus 检查当前状态是否为自动状态
func (ctx *MachineContext) IsAutoStatus() bool {
	return ctx.MachineStatus.IsAutoStatus(ctx.Status)
}

// TraceId 返回当前上下文的 trace ID
func (ctx *MachineContext) TraceId() string {
	return ctx.traceId
}

// ToExecuteStatus 将状态转换为执行状态
func (ctx *MachineContext) ToExecuteStatus() string {
	if ctx.IsFailStatus() {
		return constant.MachineExecuteStatusFailed
	}
	if ctx.IsSuccessStatus() {
		return constant.MachineExecuteStatusSuccess
	}
	return constant.MachineExecuteStatusExecuting
}

// DoNext 设置下一个状态
func (ctx *MachineContext) DoNext(nextStatus string, msg string, manuals ...info.ManualInfo) bool {
	ctx.Change = ctx.doNext(nextStatus, msg, manuals...)
	ctx.Task.Refresh(*ctx.MachineTask)
	return ctx.Change
}

// doNext 设置下一个状态，返回状态是否变更
func (ctx *MachineContext) doNext(nextStatus string, msg string, manuals ...info.ManualInfo) bool {
	curNode := ctx.CurNode()
	// 状态未变更且有等待信息，则进行等待处理
	if strings.IsBlank(nextStatus) && strings.NotBlank(msg) {
		maxWait := ctx.getMaxWait()
		curNode.MaxWait = maxWait
		if maxWait != -1 && curNode.Wait >= maxWait {
			ctx.Log().WithField("maxWait", maxWait).Warn("Task wait count exceeded maximum")
			nextStatus = constant.MachineWait
		} else {
			curNode.Wait++
		}
	} else {
		if len(manuals) != 0 {
			curNode.Manuals = append(curNode.Manuals, manuals...)
		}
	}

	// 等待二次执行
	if strings.IsBlank(nextStatus) || ctx.Status == nextStatus {
		curNode.Msg = msg
		curNode.Refresh = time.Now().UnixMilli()
		curNode.Run++
		return false
	}

	// 状态变更，执行流转
	ctx.MachineTask.DoNext(nextStatus, msg)
	return true
}

// getMaxWait 获取最大等待次数
func (ctx *MachineContext) getMaxWait() int64 {
	maxWait := ctx.Runtime.GetString(constant.MachineMaxWait)
	if strings.NotBlank(maxWait) {
		return numbers.MustInt64(maxWait)
	}
	return ctx.Config.MaxWait
}
