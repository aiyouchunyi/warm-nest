// Package mapper @Author larry
// @Date 2026/06/15
// @Desc 用户账号映射器

package mapper

import (
	"fmt"
	"sync"

	"warm-nest/pkg/tool/database/mysqls"

	"warm-nest/internal/model"
)

// UserMapper 用户账号映射器
type UserMapper struct {
	mysqls.MysqlMapper[model.User]
}

var userMapper *UserMapper
var userOnce sync.Once

// GetUserMapper 获取用户账号映射器单例
func GetUserMapper() *UserMapper {
	userOnce.Do(func() {
		userMapper = &UserMapper{
			MysqlMapper: mysqls.NewMapper[model.User](),
		}
	})
	return userMapper
}

// GetByOpenId 按微信 openid 查用户（登录入口，找不到返回 nil,nil 供上层建号）
func (m *UserMapper) GetByOpenId(openId string) (*model.User, error) {
	var user model.User
	err := m.Model().Where("open_id = ?", openId).First(&user).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by openId %s: %w", openId, err)
	}
	return &user, nil
}

// GetByUserId 按业务用户ID查用户
func (m *UserMapper) GetByUserId(userId string) (*model.User, error) {
	var user model.User
	err := m.Model().Where("user_id = ?", userId).First(&user).Error
	if err != nil {
		if mysqls.IsRecordNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by userId %s: %w", userId, err)
	}
	return &user, nil
}
