// Package tasks @Author larry
// @Date 2025/5/12 10:31
// @Desc

package dto

type TaskResultInfo struct {
	TaskId string `gorm:"comment:任务ID" json:"taskId"`
	Succ   bool   `gorm:"comment:是否执行成功" json:"succ"`
	Msg    string `gorm:"comment:执行信息" json:"msg"`
}

func NewResultInfo(taskId string, err error) TaskResultInfo {
	if err != nil {
		return TaskResultInfo{
			TaskId: taskId,
			Succ:   false,
			Msg:    err.Error(),
		}
	}
	return TaskResultInfo{
		TaskId: taskId,
		Succ:   true,
		Msg:    "成功",
	}
}
