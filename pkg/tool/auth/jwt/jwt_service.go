// Package jwt @Author larry
// File jwt_service.go
// @Date 2024/8/13 20:57:00
// @Desc
package jwt

import (
	"sync"
	"time"

	"warm-nest/pkg/app/errors"
	"warm-nest/pkg/tool/auth/code"
	"warm-nest/pkg/utils/strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

// JWTService 服务结构
type JWTService struct {
}

var jwtService *JWTService
var jwtServiceOnce sync.Once

// GetJWTService 获取JWT服务单例
func GetJWTService() *JWTService {
	jwtServiceOnce.Do(func() {
		jwtService = &JWTService{}
	})
	return jwtService
}

// Token 生成token
func (s *JWTService) Token(userId string) (string, error) {
	expireTime := time.Now().Add(time.Duration(JWTConf().Expire) * time.Second)
	claims := &jwt.RegisteredClaims{
		Subject:   userId,
		ExpiresAt: jwt.NewNumericDate(expireTime),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(JWTConf().SecretKey))
	if err != nil {
		logrus.WithError(err).Errorf("Gen token failed!")
		return "", errors.NewWithArgs(code.TokenGenFailed)
	}
	return tokenStr, nil
}

// Verify 验证token
func (s *JWTService) Verify(tokenStr string) (string, error) {
	if strings.IsBlank(tokenStr) || tokenStr == "undefined" {
		logrus.Warnf("[TOKEN-VERIFY] Token is empty or undefined! %s", tokenStr)
		return "", errors.NewWithArgs(code.TokenInvalid)
	}
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(JWTConf().SecretKey), nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return "", errors.NewWithArgs(code.TokenExpired)
			}
		}

		logrus.WithFields(logrus.Fields{
			"token": tokenStr,
		}).WithError(err).Warnf("[TOKEN-VERIFY] Parse token failed!")
		return "", errors.NewWithArgs(code.TokenInvalid)
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if ok && token.Valid {
		return claims.Subject, nil
	}
	return "", errors.NewWithArgs(code.TokenInvalid)
}
