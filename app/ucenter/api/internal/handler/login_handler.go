// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含用户登录相关的 HTTP 处理器。
package handler

import (
	"errors"
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	"mscoin_go/pkg/httputil"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// LoginHandler 处理用户登录请求。
//
// 登录流程：
//  1. 解析请求体中的登录数据（用户名、密码、验证码）
//  2. 验证验证码数据存在
//  3. 获取客户端 IP 地址
//  4. 调用 ucenter-rpc LoginClient 验证用户名密码
//  5. 返回 JWT Token 和用户信息
//
// 请求路径：POST /uc/login
// 认证要求：无需认证
//
// RPC 调用关系：
//   - LoginClient.Login -> ucenter-rpc (login pb)
//   - ucenter-rpc 负责：验证用户名密码、生成 JWT Token、记录登录日志
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 验证码是登录的必要条件，防止机器人暴力破解
		if req.Captcha == nil {
			httpx.OkJsonCtx(r.Context(), w, result.New().Deal(nil, errors.New("captcha verification failed")))
			return
		}

		// 获取客户端真实 IP，用于安全审计和风控
		// 可能是直连 IP 或经过代理后的 IP（从 X-Forwarded-For 等头部解析）
		req.IP = httputil.ClientIP(r)

		// 调用 logic 层处理登录逻辑
		resp, err := logic.NewLoginLogic(r.Context(), svcCtx).Login(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
