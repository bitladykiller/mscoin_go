// Package handler 提供 market-api 服务的 HTTP 请求处理器。
// 本文件实现交易对信息查询接口的处理器。
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

// SymbolInfoHandler 创建交易对信息查询的 HTTP 处理器。
//
// HTTP 路由: POST /market/symbol-info
//
// 请求参数 (types.MarketReq):
//   - symbol: 交易对代码，如 "BTCUSDT"
//
// 响应数据 (types.ExchangeCoinResp):
//   - 返回交易对的详细配置信息，包括精度、手续费、限制等
//
// 处理流程:
//  1. 解析表单参数到 MarketReq 结构体
//  2. 提取客户端真实 IP 地址
//  3. 调用 SymbolInfoLogic 执行业务逻辑
//  4. 使用统一响应封装器返回结果
//
// 使用场景:
//   - 交易对详情页展示
//   - 交易规则说明
//   - 交易限制提示
//
// 参数:
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回:
//   - http.HandlerFunc: HTTP 处理函数
func SymbolInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析请求参数到 MarketReq 结构体
		var req types.MarketReq
		if err := httpx.ParseForm(r, &req); err != nil {
			// 参数解析失败，返回错误响应
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 提取客户端真实 IP 地址
		req.IP = httputil.ClientIP(r)

		// 调用业务逻辑层处理请求
		resp, err := logic.NewSymbolInfoLogic(r.Context(), svcCtx).SymbolInfo(req)

		// 返回统一格式的响应
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}