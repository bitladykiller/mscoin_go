// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含发送提现验证码相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// SendWithdrawCodeHandler 处理发送提现验证码请求。
//
// 该接口用于发送提现操作所需的短信验证码，是提现流程的二次验证环节。
//
// 提现验证流程：
//  1. 用户发起提现请求前，先调用此接口获取验证码
//  2. 系统向用户已认证手机号发送验证码
//  3. 用户在提现申请时提交验证码进行验证
//
// 请求路径：POST /uc/mobile/withdraw/code
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//   - 根据用户 ID 查询会员手机号
//
// RPC 调用关系：
//   - MemberClient.FindMemberById -> ucenter-rpc (member pb)：获取会员手机号
//   - WithdrawClient.SendCode -> ucenter-rpc (withdraw pb)：发送提现验证码
//
// 安全考虑：
//   - 验证码发送到用户已认证的手机号，而非请求中指定的号码
//   - 应有发送频率限制，防止短信轰炸
func SendWithdrawCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 调用 logic 层发送验证码
		// logic 层会自动获取用户手机号并发送验证码
		resp, err := logic.NewSendWithdrawCodeLogic(r.Context(), svcCtx).SendCode()
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
