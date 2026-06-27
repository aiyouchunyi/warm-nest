// Package tasks @Author larry
// @Date 2025/5/12 10:31
// @Desc

package dto

import (
	"warm-nest/pkg/tool/machines/model/constant"
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/utils/transforms"
)

type TaskExecuteInfo struct {
	TaskId  string `gorm:"comment:任务ID" json:"taskId"`
	Status  string `gorm:"comment:状态" json:"status"`
	Time    int64  `gorm:"comment:时间" json:"time"`
	Msg     string `gorm:"comment:消息" json:"msg,omitempty"`
	Execute string `gorm:"comment:执行状态" json:"execute"`
}

func NewExecuteInfo(execute string, ctx *context.MachineContext) TaskExecuteInfo {
	executeInfo := TaskExecuteInfo{
		TaskId:  ctx.TaskId,
		Execute: execute,
	}
	transforms.TransformIgnoreNil(ctx.CurNode(), &executeInfo)
	return executeInfo
}

// IsSuccess 判断执行状态是否成功
func (e TaskExecuteInfo) IsSuccess() bool {
	return e.Execute == constant.MachineExecuteStatusSuccess
}

// IsFailed 判断执行状态是否失败
func (e TaskExecuteInfo) IsFailed() bool {
	return e.Execute == constant.MachineExecuteStatusFailed
}

// IsExecuting 判断执行状态是否执行中
func (e TaskExecuteInfo) IsExecuting() bool {
	return e.Execute == constant.MachineExecuteStatusExecuting
}
