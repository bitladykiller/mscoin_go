package middleware

import (
	"context"
	"net/http"

	"mscoin_go/pkg/auth"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type contextKey string

const userIDKey contextKey = "userId"

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("x-auth-token")
		if token == "" {
			failed := result.New()
			failed.Fail(4000, "no login")
			httpx.WriteJson(w, http.StatusOK, failed)
			return
		}

		userID, err := auth.ParseUserID(token, m.secret)
		if err != nil {
			failed := result.New()
			failed.Fail(4000, "no login")
			httpx.WriteJson(w, http.StatusOK, failed)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

func UserIDFromContext(ctx context.Context) int64 {
	value, _ := ctx.Value(userIDKey).(int64)
	return value
}

// WithUserID 将已认证的用户 ID 注入到 context 中。
//
// 为什么需要这个辅助函数：
// - API logic 在 JWT 中间件运行后从 context 读取用户身份
// - 测试和内部适配器有时需要构造该 context
// - 导出 setter 可以避免重复定义私有 key
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
