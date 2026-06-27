// Package impls @Author larry
// @Date 2026/2/12 16:20
// @Desc 状态机事件监听器接口定义

package impls

import (
	"warm-nest/pkg/tool/machines/model/context"
	"warm-nest/pkg/tool/machines/model/manual"
)

type IMachineEventListener interface {
	Name() string
	Match() []string
	Listen(ctx *context.MachineContext, event manual.ManualEvent) error
}

type MachineEventListener struct {
}

func NewEventListener() IMachineEventListener {
	return &MachineEventListener{}
}

// Name 定义事件监听器的名称
func (m *MachineEventListener) Name() string {
	return "DefaultMachineEventListener"
}

// Match 定义该事件监听器关注的事件类型，默认实现不关注任何事件
func (m *MachineEventListener) Match() []string {
	return []string{}
}

// Listen 事件监听器，监听到事件后执行相应操作，如发送通知、记录日志等
func (m *MachineEventListener) Listen(_ *context.MachineContext, _ manual.ManualEvent) error {
	return nil
}
