// Package context @Author larry
// @Date 2024/11/27 17:07
// @Desc

package task

type TaskIdentify struct {
	TaskName  string `gorm:"comment:任务名称" json:"taskName"`
	TaskField string `gorm:"comment:任务唯一标识字段" json:"taskField"`
	TaskId    string `gorm:"comment:任务唯一标识" json:"taskId"`
}
