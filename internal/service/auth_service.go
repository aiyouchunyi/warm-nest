// Package service @Author larry
// @Date 2026/06/15
// @Desc 鉴权服务（实现框架 auth.IAuthService）
//
// C 端小程序 1 期所有登录用户平权——只校验"已登录"（靠 widgets.Session 验 token），
// 不做角色/权限分级（无需 ea 那套 role/permission 体系）。故 Access 直接放行。

package service

import (
	"sync"
)

// AuthService 鉴权服务
type AuthService struct{}

var authService *AuthService
var authServiceOnce sync.Once

// GetAuthService 获取鉴权服务单例
func GetAuthService() *AuthService {
	authServiceOnce.Do(func() {
		authService = &AuthService{}
	})
	return authService
}

// Access 权限校验（C 端登录用户平权，直接放行）
func (s *AuthService) Access(userId string, method string, path string, params string) error {
	return nil
}
