// Package handler 提供 market-api 服务的 HTTP 请求处理器。
// 本文件实现行情趋势数据查询接口的处理器。
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

// SymbolThumbTrendHandler 创建行情趋势数据查询的 HTTP 处理器。
//
// HTTP 路由: POST /market/symbol-thumb-trend
//
// 请求参数:
//   - 无需额外参数，IP 地址自动获取用于地理位置优化
//
// 响应数据 ([]*types.CoinThumbResp):
//   - 返回所有交易对的行情快照列表（包含趋势数据）
//   - 每个交易对包含开盘价、收盘价、涨跌幅、成交量和 Trend 数组
//   - Trend 数组用于绘制迷你趋势图
//
// 处理流程:
//  1. 构造请求对象并自动填充客户端 IP
//  2. 调用 SymbolThumbTrendLogic 执行业务逻辑
//  3. 使用统一响应封装器返回结果
//
// 使用场景:
//   - 首页行情列表（带迷你趋势图）
//   - 市场概览页面（带趋势展示）
//   - 行情实时刷新（包含趋势变化）
//
// 与 /symbol-thumb 接口的区别:
//   - /symbol-thumb: 返回基础行情数据，不包含趋势
//   - /symbol-thumb-trend: 返回行情数据 + 趋势数组，数据量更大
//
// 参数:
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回:
//   - http.HandlerFunc: HTTP 处理函数
func SymbolThumbTrendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 构造请求对象，自动填充客户端 IP
		// 该接口不需要额外参数，IP 用于地理位置相关的数据筛选
		req := &types.MarketReq{IP: httputil.ClientIP(r)}

		// 调用业务逻辑层处理请求
		resp, err := logic.NewSymbolThumbTrendLogic(r.Context(), svcCtx).SymbolThumbTrend(req)

		// 返回统一格式的响应
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp, err))
	}
}