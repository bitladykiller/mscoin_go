// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含重置充值地址相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	"mscoin_go/pkg/httputil"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ResetAddressHandler 处理重置充值地址请求。
//
// 该接口用于重置指定币种的充值地址，主要场景：
//   - 用户需要更换充值地址
//   - 原地址出现问题（如区块链分叉）
//   - 安全考虑更换地址
//
// 请求路径：POST /uc/asset/wallet/reset-address
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 请求参数（表单提交）：
//   - Unit 或 CoinName：币种单位/名称
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//
// RPC 调用关系：
//   - AssetClient.ResetAddress -> ucenter-rpc (asset pb)
//   - ucenter-rpc 负责：生成新充值地址、更新钱包记录、记录操作日志
//
// 安全考虑：
//   - 需要记录操作 IP，用于安全审计
//   - 可能有频率限制，防止滥用
func ResetAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssetReq

		// 从表单解析请求参数
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取客户端 IP，用于安全审计
		req.IP = httputil.ClientIP(r)

		// 调用 logic 层重置地址
		resp, err := logic.NewResetAddressLogic(r.Context(), svcCtx).ResetAddress(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
