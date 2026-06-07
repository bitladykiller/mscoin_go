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

// WithUserID injects the authenticated user id into a context.
//
// Why this helper exists:
// - API logic reads user identity from context after JWT middleware runs
// - tests and internal adapters sometimes need to construct that context
// - exporting the setter avoids duplicating the private key definition
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
