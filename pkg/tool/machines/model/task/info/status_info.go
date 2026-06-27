// Package info @Author larry
// @Date 2024/5/23 09:29:00
// @Desc 任务流
package info

type StatusInfo struct {
	Status  string       `gorm:"comment:当前状态" json:"status"`
	Time    int64        `gorm:"comment:执行时间" json:"time"`
	Msg     string       `gorm:"comment:消息" json:"msg,omitempty"`
	Refresh int64        `gorm:"comment:刷新时间" json:"refresh,omitempty"`
	Run     int64        `gorm:"comment:执行次数" json:"run,omitempty"`
	Wait    int64        `gorm:"comment:等待次数" json:"wait,omitempty"`
	MaxWait int64        `gorm:"comment:最大等待次数" json:"max_wait,omitempty"`
	Manuals []ManualInfo `gorm:"comment:操作信息" json:"manuals,omitempty"`
}
