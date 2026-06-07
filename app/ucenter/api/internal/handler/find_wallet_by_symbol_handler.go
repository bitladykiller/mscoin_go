// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含按币种查询钱包相关的 HTTP 处理器。
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

// FindWalletBySymbolHandler 处理查询指定币种钱包请求。
//
// 该接口返回用户指定币种的钱包详情，主要用于：
//   - 充值页面展示充值地址和币种信息
//   - 提现页面展示余额和提现限制
//
// 请求路径：POST /uc/asset/wallet/:coinName
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// URL 参数：
//   - coinName：币种名称（如 BTC、ETH、USDT），通过 URL 路径传递
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//
// RPC 调用关系：
//   - AssetClient.FindWalletBySymbol -> ucenter-rpc (asset pb)
//   - ucenter-rpc 负责：查询指定币种钱包、返回余额和充值地址
func FindWalletBySymbolHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AssetReq

		// 从 URL 路径解析币种名称
		if err := httpx.ParsePath(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取客户端 IP，用于安全审计
		req.IP = httputil.ClientIP(r)

		// 调用 logic 层查询单个钱包
		resp, err := logic.NewFindWalletBySymbolLogic(r.Context(), svcCtx).FindWalletBySymbol(&req)
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
