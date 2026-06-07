package handler

import (
	"net/http"

	"mscoin_go/app/market/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册 market-api 暴露的 HTTP 路由。
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/coin-info", Handler: CoinInfoHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/exchange-rate/usd/:unit", Handler: UsdRateHandler(serverCtx)},
			{Method: http.MethodGet, Path: "/history", Handler: HistoryHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/symbol-info", Handler: SymbolInfoHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/symbol-thumb", Handler: SymbolThumbHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/symbol-thumb-trend", Handler: SymbolThumbTrendHandler(serverCtx)},
		},
		rest.WithPrefix("/market"),
	)
}
