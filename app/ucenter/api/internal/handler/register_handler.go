// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含用户注册相关的 HTTP 处理器。
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

// RegisterHandler 处理用户注册请求。
//
// 注册流程：
//  1. 解析请求体中的注册数据（用户名、密码、手机号、验证码、邀请码等）
//  2. 验证验证码数据存在
//  3. 获取客户端 IP 地址
//  4. 调用 ucenter-rpc RegisterClient 完成注册
//
// 请求路径：POST /uc/register/phone
// 认证要求：无需认证
//
// 支持的注册方式：
//   - 手机号注册：需要手机号、短信验证码
//   - 邀请注册：可携带邀请码建立邀请关系
//
// RPC 调用关系：
//   - RegisterClient.RegisterByPhone -> ucenter-rpc (register pb)
//   - ucenter-rpc 负责：验证短信验证码、创建会员账号、建立邀请关系
func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.Request
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 验证码是注册的必要条件，防止机器人批量注册
		if req.Captcha == nil {
			httpx.OkJsonCtx(r.Context(), w, result.New().Deal(nil, errors.New("captcha verification failed")))
			return
		}

		// 获取客户端真实 IP，用于安全审计和风控
		req.IP = httputil.ClientIP(r)

		// 调用 logic 层处理注册逻辑
		resp, err := logic.NewRegisterLogic(r.Context(), svcCtx).Register(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
