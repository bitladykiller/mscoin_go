// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含申请提现相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// WithdrawCodeHandler 处理申请提现请求。
//
// 该接口是提现流程的核心接口，执行提现申请操作。
//
// 提现流程：
//  1. 用户先调用 SendWithdrawCodeHandler 获取短信验证码
//  2. 用户填写提现信息（币种、金额、地址、资金密码、验证码）
//  3. 调用此接口提交提现申请
//  4. 系统验证通过后创建提现工单
//
// 请求路径：POST /uc/withdraw/apply/code
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 请求参数（表单提交）：
//   - Unit：币种单位（如 USDT、BTC）
//   - Address：提现目标地址
//   - Amount：提现金额
//   - Fee：矿工费
//   - JyPassword：资金密码
//   - Code：短信验证码
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//
// RPC 调用关系：
//   - WithdrawClient.WithdrawCode -> ucenter-rpc (withdraw pb)
//   - ucenter-rpc 负责：
//       - 验证资金密码
//       - 验证短信验证码
//       - 检查余额是否充足
//       - 检查提现限额
//       - 创建提现工单
//       - 冻结提现金额
//
// 安全考虑：
//   - 需要资金密码验证，防止账号被盗后资产被盗
//   - 需要短信验证码，提供二次验证
//   - 提现金额和地址需要严格校验
func WithdrawCodeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.WithdrawReq

		// 从表单解析提现请求参数
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 调用 logic 层处理提现申请
		resp, err := logic.NewWithdrawCodeLogic(r.Context(), svcCtx).WithdrawCode(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
