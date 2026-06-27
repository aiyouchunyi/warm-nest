// Package auth @Author Larry
// @Date 2024/10/12 14:14
// @Desc

package auth

import (
	"fmt"
	"sync"

	"warm-nest/pkg/app/web/session"
)

type IAuthorizeService interface {
	Token(userId string) (string, error)                                   // 生成token
	Verify(token string) (string, error)                                   // 验证token
	Access(userId string, method string, path string, params string) error // 验证用户权限
}

type IJwtService interface {
	Token(userId string) (string, error)
	Verify(token string) (string, error)
}

type IAuthService interface {
	Access(userId string, method string, path string, params string) error // 验证用户权限
}

type IUserService interface {
	SessionUser(userId string) (session.User, error) // 获取用户信息
	ClearCache(userId string)
}

type AuthorizeService struct {
	jwtService  IJwtService
	authService IAuthService
	userService IUserService
}

// Singleton instance
var authorizeService *AuthorizeService

// GetAuthorizeService returns the singleton instance of AuthorizeService
func GetAuthorizeService() *AuthorizeService {
	if authorizeService == nil {
		panic("authorize service is not start!")
	}
	return authorizeService
}

var enableMu sync.Mutex
var enableCalled bool

func EnableAuth(params ...interface{}) error {
	enableMu.Lock()
	defer enableMu.Unlock()
	if enableCalled {
		return fmt.Errorf("authorize service can only be enabled once")
	}
	if len(params) != 3 {
		return fmt.Errorf("authorize service requires 3 parameters: jwtService, authService, userService")
	}
	jwtService, ok1 := params[0].(IJwtService)
	authService, ok2 := params[1].(IAuthService)
	userService, ok3 := params[2].(IUserService)
	if !ok1 || !ok2 || !ok3 {
		return fmt.Errorf("invalid parameter types for authorize service initialization: jwtService, authService, userService")
	}
	authorizeService = &AuthorizeService{
		jwtService:  jwtService,
		authService: authService,
		userService: userService,
	}
	return nil
}

// Token 生成token
func (s *AuthorizeService) Token(userId string) (string, error) {
	return s.jwtService.Token(userId)
}

// Verify 验证token
func (s *AuthorizeService) Verify(token string) (string, session.User, error) {
	userId, err := s.jwtService.Verify(token)
	if err != nil {
		return "", session.User{}, err
	}
	userInfo, _ := s.userService.SessionUser(userId)
	return userId, userInfo, nil
}

// Access 验证用户权限
func (s *AuthorizeService) Access(userId string, method string, path string, params string) error {
	return s.authService.Access(userId, method, path, params)
}
