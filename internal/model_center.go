// Package internal @Author larry
// @Desc 模型注册中心
//
// 【新增 model 后必做】在 RegisterModel() map 里追加一行：
//
//	"<中文表说明>": model.Xxx{},
//
// key 是注释（仅展示用），value 必须是 model 结构体的零值
package internal

import (
	"warm-nest/internal/model"
)

// RegisterModel 注册数据表
func RegisterModel() map[string]interface{} {
	return map[string]interface{}{
		"用户账号表":   model.User{},
		"被守护人档案表": model.ElderProfile{},
		"收货地址簿表":  model.ShippingAddress{},
		"守护邀请表":   model.Invitation{},
		"守护关系表":   model.Guardianship{},
		"打卡记录表":   model.CheckIn{},
		"奖励任务规则表": model.RewardTask{},
		"领取记录表":   model.RewardClaim{},
		"小程序消息表":  model.Message{},
		"服务号粉丝表":  model.Fan{},
		"通知路由配置表": model.NotifyRoute{},
	}
}
