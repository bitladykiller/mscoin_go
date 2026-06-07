// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含检查登录状态相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// CheckLoginHandler 处理检查登录状态请求。
//
// 该接口用于验证 JWT Token 是否有效，主要场景：
//   - 前端页面初始化时检查用户登录状态
//   - 定期轮询检查 Token 是否过期
//
// 请求路径：POST /uc/check/login
// 认证要求：无需认证（但需要携带 x-auth-token 头部）
//
// 请求参数：
//   - 通过 x-auth-token 请求头传递 JWT Token
//
// 响应：
//   - Token 有效：返回 true
//   - Token 无效或过期：返回 false
//
// 为什么该接口无需 Auth 中间件：
//   - 该接口的目的是返回 Token 是否有效
//   - 如果使用 Auth 中间件，Token 无效时会被拦截返回错误
//   - 直接在 logic 层处理可以返回更友好的状态信息
func CheckLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从请求头获取 JWT Token
		token := r.Header.Get("x-auth-token")

		// 调用 logic 层验证 Token
		resp, err := logic.NewCheckLoginLogic(r.Context(), svcCtx).CheckLogin(token)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
