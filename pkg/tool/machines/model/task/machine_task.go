// Package tasks @Author larry
// File machine_task.go
// @Date 2024/5/8 20:52:00
// @Desc 状态机信息
package task

import (
	"time"

	"github.com/creasty/defaults"
	"github.com/sirupsen/logrus"

	"warm-nest/pkg/kinds/jsons"
	"warm-nest/pkg/kinds/variables"
	"warm-nest/pkg/tool/machines/model/task/info"
	"warm-nest/pkg/tool/machines/model/variable"
	"warm-nest/pkg/utils/slices"
	"warm-nest/pkg/utils/strings"
)

type IMachineTask interface {
	GetTask(options ...variable.Option) *MachineTask
	GetIdentify() TaskIdentify
	Refresh(task MachineTask)
}

type MachineTask struct {
	Status      string                       `gorm:"size:32;index:i_status;comment:状态" json:"status"`
	Lock        bool                         `gorm:"comment:转账任务锁定" json:"lock,omitempty"`
	LockedAt    int64                        `gorm:"index:i_lock;comment:锁定时间" json:"lockedAt,omitempty"`
	StatusInfos jsons.Array[info.StatusInfo] `gorm:"type:json;comment:状态信息路径[状态值]" json:"statusInfos,omitempty"`
	Runtime     variables.Variables          `gorm:"type:json;comment:状态机运行变量" json:"runtime,omitempty"`
}

// GetTask 获取状态机信息
func (t *MachineTask) GetTask(options ...variable.Option) *MachineTask {
	_ = defaults.Set(t)
	if t.Runtime == nil {
		t.Runtime = variables.NewVariables()
	}
	for _, option := range options {
		option(t.Runtime)
	}
	return t
}

func (t *MachineTask) LastMsg() string {
	return t.CurNode().Msg
}

// CurNode 获取当前状态信息
func (t *MachineTask) CurNode() *info.StatusInfo {
	return t.Node(-1)
}

// PreNode 获取当前状态信息
func (t *MachineTask) PreNode() *info.StatusInfo {
	return t.Node(-2)
}

// Node 获取节点信息
func (t *MachineTask) Node(index int) *info.StatusInfo {
	if t.Status == "" {
		return &info.StatusInfo{}
	}
	if len(t.StatusInfos) == 0 {
		t.StatusInfos = append(t.StatusInfos, info.StatusInfo{
			Status: t.Status,
			Time:   time.Now().UnixMilli(),
		})
	}
	if index > 0 && index <= len(t.StatusInfos) {
		return &t.StatusInfos[index-1]
	}
	if index < 0 && -index <= len(t.StatusInfos) {
		return &t.StatusInfos[len(t.StatusInfos)+index]
	}
	logrus.Warnf("node index out of range %d:%d", index, len(t.StatusInfos))
	return &info.StatusInfo{}
}

// HistoryStatuses 获取历史状态列表
func (t *MachineTask) HistoryStatuses() []string {
	return slices.Map(t.StatusInfos[0:len(t.StatusInfos)-1], func(statusInfo info.StatusInfo) string {
		return statusInfo.Status
	})
}

// DoNext 设置下一个状态
func (t *MachineTask) DoNext(nextStatus string, msg string) {
	if strings.IsBlank(nextStatus) || t.Status == nextStatus {
		return
	}
	t.StatusInfos = append(t.StatusInfos, info.StatusInfo{
		Status: nextStatus,
		Msg:    msg,
		Time:   time.Now().UnixMilli(),
	})
	t.Status = nextStatus
}

// Refresh 刷新任务信息，继承任务返回指针已经被修改
func (t *MachineTask) Refresh(task MachineTask) {
}
