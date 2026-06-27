// Package context @Author larry
// @Date 2026/1/29 14:35
// @Desc

package context

import (
	"warm-nest/pkg/utils/slices"
)

type MachineStatus struct {
	SuccStatuses   []string `json:"succStatuses"`   //成功状态列表
	FailStatuses   []string `json:"failStatuses"`   //失败状态列表
	AutoStatues    []string `json:"autoStatues"`    //自动状态列表
	ManualStatuses []string `json:"manualStatuses"` //手动状态列表
}

func NewMachineStatus(succStatuses, failStatuses, autoStatues, manualStatuses []string) MachineStatus {
	return MachineStatus{
		SuccStatuses:   succStatuses,
		FailStatuses:   failStatuses,
		AutoStatues:    autoStatues,
		ManualStatuses: manualStatuses,
	}
}

// IsHaltStatus 是否是停止状态
func (m MachineStatus) IsHaltStatus(status string) bool {
	return m.IsManualStatus(status) ||
		m.IsFinalStatus(status) ||
		!m.IsValidStatus(status)
}

func (m MachineStatus) FinalStatus() []string {
	var statuses []string
	statuses = append(statuses, m.FailStatuses...)
	statuses = append(statuses, m.SuccStatuses...)
	return slices.Unique(statuses)
}

// IsFinalStatus 是否是最终状态
func (m MachineStatus) IsFinalStatus(status string) bool {
	return slices.Contains(m.FinalStatus(), status)
}

func (m MachineStatus) IsManualStatus(status string) bool {
	return slices.Contains(m.ManualStatuses, status)
}

// IsAutoStatus 是否是自动状态
func (m MachineStatus) IsAutoStatus(status string) bool {
	return slices.Contains(m.AutoStatues, status)
}

// IsSuccessStatus 是否是成功状态
func (m MachineStatus) IsSuccessStatus(status string) bool {
	return slices.Contains(m.SuccStatuses, status)
}

// IsFailStatus 是否是失败状态
func (m MachineStatus) IsFailStatus(status string) bool {
	return slices.Contains(m.FailStatuses, status)
}

// IsValidStatus 返回所有状态列表
func (m MachineStatus) IsValidStatus(status string) bool {
	var statuses []string
	statuses = append(statuses, m.AutoStatues...)
	statuses = append(statuses, m.ManualStatuses...)
	statuses = append(statuses, m.FailStatuses...)
	statuses = append(statuses, m.SuccStatuses...)
	return slices.Contains(slices.Unique(statuses), status)
}
