package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/uc/mobile/code", Handler: SendCodeHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/uc/register/phone", Handler: RegisterHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/uc/check/login", Handler: CheckLoginHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/uc/login", Handler: LoginHandler(serverCtx)},
	})

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.Auth},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/uc/approve/security/setting", Handler: SecuritySettingHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/asset/transaction/all", Handler: FindTransactionHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/asset/wallet", Handler: FindWalletHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/asset/wallet/reset-address", Handler: ResetAddressHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/asset/wallet/:coinName", Handler: FindWalletBySymbolHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/mobile/withdraw/code", Handler: SendWithdrawCodeHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/withdraw/apply/code", Handler: WithdrawCodeHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/withdraw/record", Handler: WithdrawRecordHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/uc/withdraw/support/coin/info", Handler: QueryWithdrawCoinHandler(serverCtx)},
			}...,
		),
	)
}
