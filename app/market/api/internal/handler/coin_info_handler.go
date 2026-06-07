// Package handler 提供 market-api 服务的 HTTP 请求处理器。
// 本文件实现币种信息查询接口的处理器。
package handler

import (
	"net/http"

	"mscoin_go/app/market/api/internal/logic"
	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	"mscoin_go/pkg/httputil"
	"mscoin_go/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// CoinInfoHandler 创建币种信息查询的 HTTP 处理器。
//
// HTTP 路由: POST /market/coin-info
//
// 请求参数 (types.MarketReq):
//   - unit: 币种单位代码，如 "BTC", "ETH", "USDT"
//
// 响应数据 (types.Coin):
//   - 返回指定币种的详细信息，包括名称、状态、充提配置等
//
// 处理流程:
//  1. 解析表单参数到 MarketReq 结构体
//  2. 提取客户端真实 IP 地址
//  3. 调用 CoinInfoLogic 执行业务逻辑
//  4. 使用统一响应封装器返回结果
//
// 错误处理:
//   - 参数解析失败: 返回 400 错误
//   - RPC 调用失败: 通过统一响应封装错误信息
//
// 参数:
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回:
//   - http.HandlerFunc: HTTP 处理函数
func CoinInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析请求参数到 MarketReq 结构体
		var req types.MarketReq
		if err := httpx.ParseForm(r, &req); err != nil {
			// 参数解析失败，返回错误响应
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 提取客户端真实 IP 地址
		// 支持代理服务器场景下的 IP 获取
		req.IP = httputil.ClientIP(r)

		// 调用业务逻辑层处理请求
		resp, err := logic.NewCoinInfoLogic(r.Context(), svcCtx).CoinInfo(&req)

		// 返回统一格式的响应
		// result.New().Deal() 会自动处理成功/失败的响应格式
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}