// Package provider @Author larry
// @Date 2026/06/15
// @Desc 用户信息提供者（实现框架 auth.IUserService）

package provider

import (
	"sync"

	"warm-nest/pkg/app/web/session"
	"warm-nest/pkg/tool/caches"
	"warm-nest/pkg/utils/times"

	"warm-nest/internal/mapper"
)

// UserProvider 用户信息提供者
type UserProvider struct {
	userMapper *mapper.UserMapper
}

var userProvider *UserProvider
var userProviderOnce sync.Once

// GetUserProvider 获取用户信息提供者单例
func GetUserProvider() *UserProvider {
	userProviderOnce.Do(func() {
		userProvider = &UserProvider{
			userMapper: mapper.GetUserMapper(),
		}
	})
	return userProvider
}

// SessionUser 取登录用户信息（带缓存，token 校验后回填进 ctx）
func (p *UserProvider) SessionUser(userId string) (session.User, error) {
	cacheKey := "warm-nest:session-user:" + userId
	return caches.GetOrLoad[session.User](cacheKey, times.TenMinuteInSec, func() (interface{}, error) {
		user, err := p.userMapper.GetByUserId(userId)
		if err != nil {
			return session.User{}, err
		}
		if user == nil {
			return session.User{}, nil
		}
		return session.User{
			UserId: user.UserId,
			Status: user.Status,
			Name:   user.Nickname,
		}, nil
	})
}

// ClearCache 清除用户缓存
func (p *UserProvider) ClearCache(userId string) {
	caches.Delete("warm-nest:session-user:" + userId)
}
