// Package service @Author larry
// @Date 2026/06/18
// @Desc 服务号粉丝服务（PRD §8.2 关注/取关事件维护关注态 + 关注判定）
//
// 关注态的唯一数据源是服务号「关注/取关」事件回调（小程序端无 API 可查）。
// 跨主体锚点是 UnionId：判某小程序用户是否关注 = 取其 User.UnionId → 查 Fan.Subscribed。

package service

import (
	"fmt"
	"sync"

	"warm-nest/internal/mapper"
	"warm-nest/internal/model"
)

// FanService 服务号粉丝服务
type FanService struct {
	fanMapper  *mapper.FanMapper
	userMapper *mapper.UserMapper
}

var fanService *FanService
var fanServiceOnce sync.Once

// GetFanService 获取服务号粉丝服务单例
func GetFanService() *FanService {
	fanServiceOnce.Do(func() {
		fanService = &FanService{
			fanMapper:  mapper.GetFanMapper(),
			userMapper: mapper.GetUserMapper(),
		}
	})
	return fanService
}

// ApplySubscribeEvent 落库一次关注/取关事件（subscribed=true 关注，false 取关）。
// 按 unionId 不存在则建、存在则更新关注态与时刻；幂等（同一事件重投结果一致）。
func (s *FanService) ApplySubscribeEvent(unionId, officialOpenId string, subscribed bool, nowMs int64) error {
	if unionId == "" {
		return fmt.Errorf("apply subscribe event: empty unionId")
	}
	fan, err := s.fanMapper.GetByUnionId(unionId)
	if err != nil {
		return fmt.Errorf("apply subscribe event get fan: %w", err)
	}
	isNew := fan == nil
	if isNew {
		fan = &model.Fan{UnionId: unionId}
	}
	fan.OfficialOpenId = officialOpenId
	fan.Subscribed = subscribed
	if subscribed {
		fan.SubscribeAt = nowMs
	} else {
		fan.UnsubscribeAt = nowMs
	}
	if isNew {
		return s.fanMapper.Create(fan)
	}
	return s.fanMapper.Save(fan)
}

// IsSubscribedByUserId 判某小程序用户当前是否关注服务号（查不到关注记录视为未关注）。
// 用于 §8.3 未打卡提醒第一段（仅老人已关注才推本人）与 §8.2 关注引导判定。
func (s *FanService) IsSubscribedByUserId(userId string) (bool, error) {
	user, err := s.userMapper.GetByUserId(userId)
	if err != nil {
		return false, fmt.Errorf("is subscribed get user %s: %w", userId, err)
	}
	if user == nil || user.UnionId == "" {
		return false, nil
	}
	fan, err := s.fanMapper.GetByUnionId(user.UnionId)
	if err != nil {
		return false, fmt.Errorf("is subscribed get fan: %w", err)
	}
	return fan != nil && fan.Subscribed, nil
}
