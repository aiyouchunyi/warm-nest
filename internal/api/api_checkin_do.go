// Package api @Author larry
// @Date 2026/06/15
// @Desc 打卡相关接口

package api

import (
	"warm-nest/pkg/app/web/session"

	"warm-nest/internal/service"
)

// DoCheckInReq 打卡请求（老人端，需登录）
type DoCheckInReq struct {
	session.Session
	PhotoUrl string `json:"photoUrl" validate:"required"` // 上传接口返回的图片 URL
	Weather  string `json:"weather"`                      // 前端带入的天气文案
	City     string `json:"city"`                         // 所在城市
}

// DoCheckIn 老人打卡
func (a *ApiCheckIn) DoCheckIn(req DoCheckInReq) (interface{}, error) {
	return service.GetCheckInService().DoCheckIn(req.ReqUser, req.PhotoUrl, req.Weather, req.City)
}

// CheckInStatusReq 今日打卡状态请求（公用：子女传 elderUserId 看老人今日，老人不传看自己）
type CheckInStatusReq struct {
	session.Session
	ElderUserId string `form:"elderUserId"` // 可选：看某老人的（守护校验），不传=自己
}

// TodayStatus 今日是否已打卡（公用：子女传 elderUserId 看老人今日状态）
func (a *ApiCheckIn) TodayStatus(req CheckInStatusReq) (interface{}, error) {
	elderUserId, err := service.GetFamilyService().ResolveElder(req.ReqUser, req.ElderUserId)
	if err != nil {
		return nil, err
	}
	checked, err := service.GetCheckInService().TodayChecked(elderUserId)
	if err != nil {
		return nil, err
	}
	return map[string]bool{"checked": checked}, nil
}

// MonthCheckInReq 月份打卡记录请求（公用：子女传 elderUserId 看老人的，老人不传看自己的）
type MonthCheckInReq struct {
	session.Session
	YearMonth   string `form:"yearMonth" validate:"required"` // YYYY-MM
	ElderUserId string `form:"elderUserId"`                   // 可选：看某老人的（守护校验），不传=自己
}

// MonthCheckIns 查某月打卡记录（日历/月份详情，问题8 公用）
func (a *ApiCheckIn) MonthCheckIns(req MonthCheckInReq) (interface{}, error) {
	elderUserId, err := service.GetFamilyService().ResolveElder(req.ReqUser, req.ElderUserId)
	if err != nil {
		return nil, err
	}
	return service.GetCheckInService().MonthCheckIns(elderUserId, req.YearMonth)
}

// MonthOverviewReq 打卡月份记录页请求（公用：子女传 elderUserId 看老人的，老人不传看自己的）
type MonthOverviewReq struct {
	session.Session
	ElderUserId string `form:"elderUserId"` // 可选：看某老人的（守护校验），不传=自己
}

// MonthOverview 打卡月份记录页：累计打卡/已领鸡蛋/待领月份 + 逐月进度（问题10，问题7 公用）
func (a *ApiCheckIn) MonthOverview(req MonthOverviewReq) (interface{}, error) {
	elderUserId, err := service.GetFamilyService().ResolveElder(req.ReqUser, req.ElderUserId)
	if err != nil {
		return nil, err
	}
	return service.GetRewardService().MonthOverview(elderUserId)
}

// CheckInDetailReq 单日打卡详情请求（公用·模式B：只传 checkInId，归属由记录定）
type CheckInDetailReq struct {
	session.Session
	CheckInId string `form:"checkInId" validate:"required"`
}

// CheckInDetail 按 checkInId 查单日打卡详情（问题4，老人/子女公用，守护校验防越权）
func (a *ApiCheckIn) CheckInDetail(req CheckInDetailReq) (interface{}, error) {
	return service.GetCheckInService().GetCheckInDetail(req.ReqUser, req.CheckInId)
}
