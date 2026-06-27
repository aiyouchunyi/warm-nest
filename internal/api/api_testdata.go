// Package api @Author larry
// @Date 2026/06/19
// @Desc 测试数据接口（供前端自助造数据；临时能力，测完手动删本文件 + controller 注册）
//
// 文件名用 api_testdata.go 而非 api_test.go：Go 把 *_test.go 视为测试文件、不编译进正式包，
// 用 testdata 资源名避开该后缀语义。无鉴权、全 POST（造/改数据），所有环境可用。

package api

import (
	"sync"

	"warm-nest/internal/service"
)

// ApiTestData 测试数据接口
type ApiTestData struct{}

var apiTestData *ApiTestData
var apiTestDataOnce sync.Once

// GetApiTestData 获取测试数据接口单例
func GetApiTestData() *ApiTestData {
	apiTestDataOnce.Do(func() {
		apiTestData = &ApiTestData{}
	})
	return apiTestData
}

// SeedCheckInsReq 造连续打卡请求（无鉴权）
type SeedCheckInsReq struct {
	ElderUserId string `json:"elderUserId" validate:"required"` // 给谁造打卡
	EndDate     string `json:"endDate"`                         // 截止日 YYYY-MM-DD，空=今天
	Days        int    `json:"days"`                            // 连续天数
}

// SeedCheckIns 造「截至 endDate 往前连续 days 天」的打卡记录
func (a *ApiTestData) SeedCheckIns(req SeedCheckInsReq) (interface{}, error) {
	return service.GetTestService().SeedContinuousCheckIns(req.ElderUserId, req.EndDate, req.Days)
}

// SeedRewardTasksReq 初始化奖励规则种子请求（无入参）
type SeedRewardTasksReq struct{}

// SeedRewardTasks 初始化默认奖励任务规则（幂等）
func (a *ApiTestData) SeedRewardTasks(_ SeedRewardTasksReq) (interface{}, error) {
	return service.GetTestService().SeedRewardTasks()
}

// SeedRewardClaimsReq 直插领取记录请求（无鉴权）
type SeedRewardClaimsReq struct {
	ElderUserId string `json:"elderUserId" validate:"required"` // 归属老人
	Status      string `json:"status"`                          // PENDING/CLAIMED/SHIPPED/SIGNED，空=PENDING
	Count       int    `json:"count"`                           // 造几条，<=0 取 1
	Quantity    int    `json:"quantity"`                        // 每条奖励数量，<=0 取 1
}

// SeedRewardClaims 直插 N 条指定状态的领取记录（绕过打卡评估，专用于测列表多状态/物流展示）
func (a *ApiTestData) SeedRewardClaims(req SeedRewardClaimsReq) (interface{}, error) {
	return service.GetTestService().SeedRewardClaims(req.ElderUserId, req.Status, req.Count, req.Quantity)
}

// TriggerRemindReq 手动触发未打卡提醒轮询请求（无鉴权，测试用）
type TriggerRemindReq struct {
	At string `json:"at"` // 可选：模拟触发时刻 HH:mm（按 Asia/Shanghai 当天），空=真实当前时刻
}

// TriggerRemind 手动触发一次未打卡提醒轮询（模拟定时任务），便于联调验证两段式推送
func (a *ApiTestData) TriggerRemind(req TriggerRemindReq) (interface{}, error) {
	return service.GetTestService().TriggerRemindPolling(req.At)
}

// PublishMenuReq 发布服务号默认菜单请求（无入参）
type PublishMenuReq struct{}

// PublishMenu 发布服务号 1.0 默认底部菜单（打卡/我的，均跳小程序入口页 pages/login/index）
func (a *ApiTestData) PublishMenu(_ PublishMenuReq) (interface{}, error) {
	return service.GetMenuService().PublishDefaultMenu()
}

// WipeAllReq 全清数据库请求（无入参；⚠️不可逆，仅测试通道）
type WipeAllReq struct{}

// WipeAll 清空全部业务表 + 账号表（问题10），返回各表清理条数
func (a *ApiTestData) WipeAll(_ WipeAllReq) (interface{}, error) {
	return service.GetTestService().WipeAll()
}

// SeedGuardianshipReq 造守护关系请求（无鉴权）
type SeedGuardianshipReq struct {
	GuardianUserId string `json:"guardianUserId" validate:"required"` // 子女
	ElderUserId    string `json:"elderUserId" validate:"required"`    // 老人
	Relation       string `json:"relation"`                           // 称呼，空=DAD
}

// SeedGuardianship 造「子女→老人」ACTIVE 守护关系 + 双方账号 + 老人档案 + 默认收货地址
func (a *ApiTestData) SeedGuardianship(req SeedGuardianshipReq) (interface{}, error) {
	return service.GetTestService().SeedGuardianship(req.GuardianUserId, req.ElderUserId, req.Relation)
}

// SeedPendingInvitationReq 造待接受邀请请求（无鉴权）
type SeedPendingInvitationReq struct {
	GuardianUserId string `json:"guardianUserId" validate:"required"` // 发起邀请的子女
	Relation       string `json:"relation"`                           // 称呼，空=DAD
}

// SeedPendingInvitation 造一条 PENDING 邀请 + 子女账号（带头像/手机号），供测问题1/4/5
func (a *ApiTestData) SeedPendingInvitation(req SeedPendingInvitationReq) (interface{}, error) {
	return service.GetTestService().SeedPendingInvitation(req.GuardianUserId, req.Relation)
}
