// Package middleware 定义了 exchange-api 的 HTTP 中间件。
// 提供认证中间件用于验证 JWT 令牌并提取用户信息。
package middleware

import (
	"context"
	"net/http"

	"mscoin_go/pkg/auth"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// contextKey 定义上下文键的类型，确保类型安全。
type contextKey string

// userIDKey 是存储用户 ID 的上下文键。
const userIDKey contextKey = "userId"

// AuthMiddleware 是 JWT 认证中间件。
// 负责验证请求头中的 JWT 令牌，并将用户 ID 存入请求上下文。
type AuthMiddleware struct {
	// secret 是 JWT 签名验证密钥。
	secret string
}

// NewAuthMiddleware 创建认证中间件实例。
func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

// Handle 实现中间件处理函数。
// 处理流程：
// 1. 从请求头获取 x-auth-token
// 2. 验证令牌有效性
// 3. 解析用户 ID 并存入上下文
// 4. 调用下一个处理器
// 如果令牌缺失或无效，返回 4000 错误码表示未登录。
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取认证令牌
		token := r.Header.Get("x-auth-token")
		if token == "" {
			failed := result.New()
			failed.Fail(4000, "no login")
			httpx.WriteJson(w, http.StatusOK, failed)
			return
		}

		// 解析令牌获取用户 ID
		userID, err := auth.ParseUserID(token, m.secret)
		if err != nil {
			failed := result.New()
			failed.Fail(4000, "no login")
			httpx.WriteJson(w, http.StatusOK, failed)
			return
		}

		// 将用户 ID 存入上下文，供后续处理器使用
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext 从上下文中提取用户 ID。
// 如果上下文中不存在用户 ID，返回 0。
func UserIDFromContext(ctx context.Context) int64 {
	value, _ := ctx.Value(userIDKey).(int64)
	return value
}
