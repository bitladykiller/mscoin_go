package handler

import (
	"net/http"

	"mscoin_go/app/exchange/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

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
