package handler

import (
	"net/http"

	"mscoin_go/app/exchange/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册 exchange-api 的所有路由处理器。
// 所有路由都需要经过认证中间件验证，路由前缀为 "/exchange"。
// 注册的路由包括：
// - POST /exchange/order/add：新增订单
// - POST /exchange/order/current：查询当前订单
// - POST /exchange/order/history：查询历史订单
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
