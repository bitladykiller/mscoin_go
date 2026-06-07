// Package handler 提供 market-api 服务的 HTTP 请求处理器。
// 本文件实现法币汇率查询接口的处理器。
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

// UsdRateHandler 创建法币汇率查询的 HTTP 处理器。
//
// HTTP 路由: POST /market/exchange-rate/usd/:unit
//
// 请求参数:
//   - unit: 法币单位代码（路径参数），如 "CNY", "EUR", "JPY", "KRW"
//   - ip: 客户端 IP（可选，自动填充）
//
// 响应数据:
//   - rate: 指定法币对 USD 的汇率值
//   - 例如 CNY 返回 {"rate": 7.24}，表示 1 USD = 7.24 CNY
//
// 处理流程:
//  1. 解析路径参数 unit 到 RateRequest 结构体
//  2. 提取客户端真实 IP 地址
//  3. 调用 UsdRateLogic 执行业务逻辑
//  4. 使用统一响应封装器返回汇率值
//
// 使用场景:
//   - 法币资产价值换算
//   - 多币种价格展示
//   - 汇率实时更新
//
// 注意事项:
//   - 汇率数据由后端 rate-rpc 服务提供
//   - 支持基于 IP 地理位置的汇率自动选择
//
// 参数:
//   - svcCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 返回:
//   - http.HandlerFunc: HTTP 处理函数
func UsdRateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 解析路径参数到 RateRequest 结构体
		// unit 参数从 URL 路径中提取
		var req types.RateRequest
		if err := httpx.ParsePath(r, &req); err != nil {
			// 参数解析失败，返回错误响应
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 提取客户端真实 IP 地址
		// 用于基于地理位置的汇率查询优化
		req.IP = httputil.ClientIP(r)

		// 调用业务逻辑层处理请求
		resp, err := logic.NewUsdRateLogic(r.Context(), svcCtx).UsdRate(&req)

		// 返回统一格式的响应
		// 注意：这里只返回 resp.Rate 部分，保持与旧 API 的兼容性
		httpx.OkJsonCtx(r.Context(), w, result.New().Deal(resp.Rate, err))
	}
}