// Package middleware 提供 ucenter-api 服务的 HTTP 中间件。
//
// 该包主要提供 JWT 认证中间件，用于保护需要用户登录才能访问的接口。
//
// 认证流程：
//  1. 从请求头获取 x-auth-token
//  2. 解析 Token 获取用户 ID
//  3. 将用户 ID 注入到 context 中供后续逻辑使用
//  4. 如果 Token 无效或缺失，返回 4000 错误码（未登录）
//
// 使用方式：
//   - 在 routes.go 中通过 rest.WithMiddlewares 注册需要认证的路由组
//   - 在 logic 层通过 middleware.UserIDFromContext 获取用户 ID
package middleware

import (
	"context"
	"net/http"

	"mscoin_go/pkg/auth"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// contextKey 是 context.Value 的键类型。
// 使用自定义类型可以避免键冲突。
type contextKey string

// userIDKey 是存储用户 ID 的 context 键。
// 通过 UserIDFromContext 函数获取值，通过 WithUserID 函数设置值。
const userIDKey contextKey = "userId"

// AuthMiddleware 是 JWT 认证中间件结构。
//
// 该中间件负责验证请求中的 JWT Token，并将解析出的用户 ID 注入到 context。
// 这是保护用户相关接口的第一道防线。
type AuthMiddleware struct {
	// secret 是 JWT Token 的签名密钥。
	// 必须与 ucenter-rpc 服务生成 Token 时使用的密钥一致。
	secret string
}

// NewAuthMiddleware 创建新的认证中间件实例。
//
// 参数：
//   - secret：JWT 签名密钥，从配置文件读取
//
// 返回：
//   - *AuthMiddleware：认证中间件实例
func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

// Handle 是中间件的处理函数，实现 rest.Middleware 接口。
//
// 认证流程详解：
//  1. 从请求头 x-auth-token 获取 JWT Token
//  2. 如果 Token 为空，返回错误码 4000（未登录）
//  3. 使用 auth.ParseUserID 解析 Token 获取用户 ID
//  4. 如果解析失败，返回错误码 4000（未登录）
//  5. 将用户 ID 注入 context，调用下一个处理器
//
// 为什么选择 4000 作为错误码：
//   - 前端约定使用 4000 表示"未登录"状态
//   - 区别于 HTTP 401 状态码，保持响应体格式一致
//
// 为什么不在 Token 无效时返回更详细的错误：
//   - 安全考虑：不暴露 Token 具体问题（过期、签名错误等）
//   - 统一处理：前端只需判断是否需要重新登录
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取 Token，x-auth-token 是前后端约定的认证头部名称
		token := r.Header.Get("x-auth-token")
		if token == "" {
			failed := result.New()
			failed.Fail(4000, "no login")
			httpx.WriteJson(w, http.StatusOK, failed)
			return
		}

		// 解析 Token 获取用户 ID
		// ParseUserID 内部会验证签名、检查过期时间等
		userID, err := auth.ParseUserID(token, m.secret)
		if err != nil {
			failed := result.New()
			failed.Fail(4000, "no login")
			httpx.WriteJson(w, http.StatusOK, failed)
			return
		}

		// 将用户 ID 注入 context，供后续 logic 层使用
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext 从 context 中获取已认证的用户 ID。
//
// 该函数是 middleware 包的主要导出函数，供 logic 层获取当前登录用户身份。
//
// 使用示例：
//
//	userID := middleware.UserIDFromContext(ctx)
//	if userID == 0 {
//	    // 未登录或 Token 无效
//	}
//
// 参数：
//   - ctx：HTTP 请求的 context，由中间件注入用户 ID
//
// 返回：
//   - int64：用户 ID，如果未认证则返回 0
func UserIDFromContext(ctx context.Context) int64 {
	value, _ := ctx.Value(userIDKey).(int64)
	return value
}

// WithUserID 将已认证的用户 ID 注入到 context 中。
//
// 为什么需要这个辅助函数：
//   - API logic 在 JWT 中间件运行后从 context 读取用户身份
//   - 测试和内部适配器有时需要构造该 context
//   - 导出 setter 可以避免重复定义私有 key
//
// 使用场景：
//   - 单元测试中模拟已认证用户
//   - 内部服务调用时传递用户身份
//
// 参数：
//   - ctx：原始 context
//   - userID：要注入的用户 ID
//
// 返回：
//   - context.Context：包含用户 ID 的新 context
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
