// Package auth 包含 API 服务间共享的身份验证辅助工具。
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateUserToken 创建 MSCoin API 使用的传统 JWT。
//
// Token 特意保留历史 claim 名称，以便在迁移期间所有 API 中间件和现有客户端保持兼容。
func GenerateUserToken(secret string, issuedAt time.Time, expireSeconds int64, userID int64) (string, error) {
	claims := jwt.MapClaims{
		"exp":    issuedAt.Unix() + expireSeconds,
		"iat":    issuedAt.Unix(),
		"userId": userID,
	}

	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secret))
}

// ParseUserID 从 JWT token 中提取传统的 `userId` claim。
//
// 为什么需要这个辅助函数：
// - 旧项目将登录状态存储在 `x-auth-token` 中
// - 多个 API 服务需要相同的 claim 解析行为
// - 集中 token 解析可以避免各服务之间出现微妙的偏差
func ParseUserID(tokenString string, secret string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	userID, ok := claims["userId"].(float64)
	if !ok {
		return 0, errors.New("userId claim missing")
	}

	expireAt, ok := claims["exp"].(float64)
	if !ok {
		return 0, errors.New("exp claim missing")
	}
	if int64(expireAt) <= time.Now().Unix() {
		return 0, errors.New("token expired")
	}

	return int64(userID), nil
}
