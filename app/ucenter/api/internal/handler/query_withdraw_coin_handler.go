// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该文件包含查询可提现币种信息相关的 HTTP 处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/logic"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// QueryWithdrawCoinHandler 处理查询可提现币种信息请求。
//
// 该接口返回提现页面所需的聚合信息，包括：
//   - 所有币种的提现配置（限额、手续费等）
//   - 用户各币种钱包余额
//   - 用户已保存的提现地址列表
//
// 该接口是提现页面初始化的核心接口，一次性获取所有展示所需数据。
//
// 请求路径：POST /uc/withdraw/support/coin/info
// 认证要求：需要 JWT Token（通过 Auth 中间件验证）
//
// 用户身份获取：
//   - 通过 middleware.UserIDFromContext 从 context 获取用户 ID
//
// RPC 调用关系：
//   - MarketClient.FindAllCoin -> market-rpc (market pb)：获取所有币种配置
//   - AssetClient.FindWallet -> ucenter-rpc (asset pb)：获取用户钱包余额
//   - WithdrawClient.FindAddressByCoinId -> ucenter-rpc (withdraw pb)：获取用户提现地址簿
//
// 数据聚合逻辑：
//   1. 从 market-rpc 获取币种列表和配置
//   2. 从 ucenter-rpc 获取用户钱包余额
//   3. 从 ucenter-rpc 获取各币种的提现地址簿
//   4. 按币种聚合上述数据，生成 WithdrawWalletInfo 列表
func QueryWithdrawCoinHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 调用 logic 层查询可提现币种信息
		// logic 层会聚合多个 RPC 服务的数据
		resp, err := logic.NewQueryWithdrawCoinLogic(r.Context(), svcCtx).QueryWithdrawCoin()
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}
