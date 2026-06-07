// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含发送短信验证码相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// SendCodeHandler 处理发送短信验证码请求。
//
// 该接口用于发送短信验证码，主要场景：
//   - 用户注册时验证手机号
//   - 用户找回密码时验证身份
//
// 请求路径：POST /uc/mobile/code
// 认证要求：无需认证
//
// 请求参数：
//   - Phone：接收验证码的手机号
//   - Country：国家代码（支持国际短信）
//
// RPC 调用关系：
//   - RegisterClient.SendCode -> ucenter-rpc (register pb)
//   - ucenter-rpc 负责：验证手机号格式、调用短信服务发送验证码、记录发送日志
//
// 安全考虑：
//   - 应有发送频率限制，防止短信轰炸
//   - 应有每日发送次数限制
func SendCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CodeRequest
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		resp, err := logic.NewSendCodeLogic(r.Context(), svcCtx).SendCode(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
