// Package handler 定义了 exchange-api 的 HTTP 请求处理器。
// 每个处理器负责解析请求、调用业务逻辑、返回响应。
//
// 路由架构说明：
// - 所有路由都需要经过认证中间件验证 JWT 令牌
// - 路由前缀为 "/exchange"
// - 处理器通过 ServiceContext 调用 exchange-rpc 服务完成业务操作
//
// 与 exchange-rpc 的调用关系：
// - AddHandler -> exchange-rpc OrderClient.Add() 创建订单
// - CurrentHandler -> exchange-rpc OrderClient.FindOrderCurrent() 查询当前委托
// - HistoryHandler -> exchange-rpc OrderClient.FindOrderHistory() 查询历史订单
package handler

import (
	"net/http"

	"mscoin_go/app/exchange/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册 exchange-api 的所有路由处理器。
// 所有路由都需要经过认证中间件验证，路由前缀为 "/exchange"。
//
// 注册的路由包括：
// - POST /exchange/order/add：新增订单，调用 exchange-rpc 的 Add 方法
// - POST /exchange/order/current：查询当前订单，调用 exchange-rpc 的 FindOrderCurrent 方法
// - POST /exchange/order/history：查询历史订单，调用 exchange-rpc 的 FindOrderHistory 方法
//
// 认证流程：
// 1. 请求到达时，先经过 Auth 中间件验证 JWT 令牌
// 2. 验证通过后，用户 ID 存入请求上下文
// 3. 后续处理器从上下文获取用户 ID 进行业务操作
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.Auth},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/order/add", Handler: AddHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/order/current", Handler: CurrentHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/order/history", Handler: HistoryHandler(serverCtx)},
			}...,
		),
		rest.WithPrefix("/exchange"),
	)
}
