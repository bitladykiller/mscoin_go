// Package handler 提供 market-api 服务的 HTTP 请求处理器。
// 本文件实现 K 线历史数据查询接口的处理器。
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

// HistoryHandler 创建 K 线历史数据查询的 HTTP 处理器。
//
// HTTP 路由: GET /market/history
//
// 请求参数 (types.MarketReq):
//   - symbol: 交易对代码，如 "BTCUSDT"
//   - from: 起始时间戳（毫秒）
//   - to: 结束时间戳（毫秒）
//   - resolution: K 线周期，如 "1", "5", "15", "30", "60", "1D"
//
// 响应数据 (types.HistoryKline):
//   - list: K 线数据数组，每项格式为 [时间戳, 开盘价, 最高价, 最低价, 收盘价, 成交量]
//
// 处理流程:
//  1. 解析查询参数到 MarketReq 结构体
//  2. 提取客户端真实 IP 地址
//  3. 调用 HistoryLogic 执行业务逻辑
//  4. 使用统一响应封装器返回结果（仅返回 list 部分）
//
// 使用场景:
//   - K 线图表数据加载
//   - 历史行情分析
//   - 技术指标计算
//
// 参数:
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回:
//   - http.HandlerFunc: HTTP 处理函数
func HistoryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析查询参数到 MarketReq 结构体
		var req types.MarketReq
		if err := httpx.ParseForm(r, &req); err != nil {
			// 参数解析失败，返回错误响应
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 提取客户端真实 IP 地址
		req.IP = httputil.ClientIP(r)

		// 调用业务逻辑层处理请求
		resp, err := logic.NewHistoryLogic(r.Context(), svcCtx).History(&req)

		// 返回统一格式的响应
		// 注意：这里只返回 resp.List 部分，保持与旧 API 的兼容性
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp.List, err))
	}
}