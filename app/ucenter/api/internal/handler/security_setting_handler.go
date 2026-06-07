// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含会员安全设置相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// SecuritySettingHandler 处理查询会员安全设置请求。
//
// 该接口返回用户的安全认证状态信息，包括：
//   - 实名认证状态（是否已实名、审核状态）
//   - 手机认证状态（已认证手机号）
//   - 邮箱认证状态（已认证邮箱）
//   - 资金密码设置状态
//   - 账户验证状态（银行卡、支付宝、微信绑定）
//
// 请求路径：POST /uc/approve/security/setting
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//   - 用户 ID 由 Auth 中间件从 JWT Token 解析并注入
//
// RPC 调用关系：
//   - MemberClient.FindMemberById -> ucenter-rpc (member pb)
//   - ucenter-rpc 负责：查询会员详细信息（实名、手机、邮箱、资金密码等）
func SecuritySettingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ApproveReq

		// 调用 logic 层查询安全设置
		resp, err := logic.NewSecuritySettingLogic(r.Context(), svcCtx).FindSecuritySetting(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
