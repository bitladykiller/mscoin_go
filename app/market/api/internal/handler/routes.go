// Package handler 提供 market-api 服务的 HTTP 请求处理器。
// 每个处理器负责解析请求参数、调用对应的业务逻辑、返回响应结果。
//
// 架构说明：
//   - handler 层：负责 HTTP 协议相关处理（参数解析、响应封装）
//   - logic 层：负责业务逻辑处理（通过 RPC 调用后端服务）
//   - handler 通过 ServiceContext 获取 RPC 客户端依赖
//
// 处理流程：
//  1. 接收 HTTP 请求
//  2. 解析请求参数到 types 结构体
//  3. 提取客户端 IP 等元信息
//  4. 调用对应的 logic 层方法
//  5. 封装响应结果返回给客户端
//
// 注意事项：
//   - 使用统一的结果封装器 (result.New().Deal()) 处理响应
//   - 使用 httputil.ClientIP() 获取真实客户端 IP
//   - 所有 handler 返回 JSON 格式响应
package handler

import (
	"net/http"

	"mscoin_go/app/market/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册 market-api 暴露的 HTTP 路由。
// 该函数将所有路由处理器注册到 go-zero REST 服务器。
//
// 路由前缀: /market
//
// 注册的路由列表:
//
//	POST /market/coin-info           - 查询币种信息
//	POST /market/exchange-rate/usd/:unit - 查询法币汇率
//	GET  /market/history             - 查询 K 线历史数据
//	POST /market/symbol-info         - 查询交易对信息
//	POST /market/symbol-thumb        - 查询行情缩略图
//	POST /market/symbol-thumb-trend  - 查询行情趋势数据
//
// 参数:
//   - server: go-zero REST 服务器实例
//   - serverCtx: 服务上下文，包含 RPC 客户端等依赖
//
// 使用示例:
//
//	server := rest.MustNewServer(c.RestConf)
//	ctx := svc.NewServiceContext(c)
//	handler.RegisterHandlers(server, ctx)
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			// 币种信息查询接口
			// 根据 unit 参数返回指定币种的详细信息
			{Method: http.MethodPost, Path: "/coin-info", Handler: CoinInfoHandler(serverCtx)},

			// 汇率查询接口
			// 查询指定法币单位对 USD 的实时汇率
			{Method: http.MethodPost, Path: "/exchange-rate/usd/:unit", Handler: UsdRateHandler(serverCtx)},

			// K 线历史数据查询接口
			// 根据 symbol, from, to, resolution 参数返回历史 K 线数据
			{Method: http.MethodGet, Path: "/history", Handler: HistoryHandler(serverCtx)},

			// 交易对信息查询接口
			// 根据 symbol 参数返回交易对的详细配置信息
			{Method: http.MethodPost, Path: "/symbol-info", Handler: SymbolInfoHandler(serverCtx)},

			// 行情缩略图数据接口
			// 返回所有交易对的行情快照，用于首页展示
			{Method: http.MethodPost, Path: "/symbol-thumb", Handler: SymbolThumbHandler(serverCtx)},

			// 行情趋势数据接口
			// 返回带有趋势数据的行情快照，用于绘制迷你趋势图
			{Method: http.MethodPost, Path: "/symbol-thumb-trend", Handler: SymbolThumbTrendHandler(serverCtx)},
		},
		rest.WithPrefix("/market"), // 所有路由添加 /market 前缀
	)
}