// Package machines @Author Larry
// @Date 2024/10/24 18:58
// @Desc

package constant

const (
	MachineInit    = "INIT"    // 任务初始化-INIT状态
	MachineSuccess = "SUCCESS" // 任务成功-FIN状态
	MachineFailed  = "FAILED"  // 任务失败-FIN状态
	MachineCancel  = "CANCEL"  // 任务手动状态取消-Manual状态
	MachineWait    = "WAIT"    // 任务自动重试超次数等待-Manual状态
)
