// Package handler 提供 ucenter-api 服务的 HTTP 请求处理器。
//
// 该包包含所有 HTTP 接口的处理器（Handler）和路由注册。
// Handler 层的职责：
//   - 解析 HTTP 请求参数
//   - 调用 logic 层执行业务逻辑
//   - 格式化并返回 HTTP 响应
//
// 路由分组：
//   - 公开路由：无需认证（注册、登录、发送验证码、检查登录状态）
//   - 认证路由：需要 JWT Token 验证（安全设置、钱包、交易、提现等）
//
// 与 RPC 服务调用关系：
//   - 公开路由 -> ucenter-rpc (register, login)
//   - 认证路由 -> ucenter-rpc (member, asset, withdraw), market-rpc (market)
package handler

import (
	"net/http"

	"mscoin_go/app/ucenter/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterHandlers 注册所有 HTTP 路由处理器。
//
// 路由分为两组：
//
// 1. 公开路由（无需认证）：
//   - POST /uc/mobile/code - 发送短信验证码
//   - POST /uc/register/phone - 手机号注册
//   - POST /uc/check/login - 检查登录状态
//   - POST /uc/login - 用户登录
//
// 2. 认证路由（需要 JWT Token）：
//   - POST /uc/approve/security/setting - 查询安全设置
//   - POST /uc/asset/transaction/all - 查询交易记录
//   - POST /uc/asset/wallet - 查询所有钱包
//   - POST /uc/asset/wallet/reset-address - 重置充值地址
//   - POST /uc/asset/wallet/:coinName - 查询单个钱包
//   - POST /uc/mobile/withdraw/code - 发送提现验证码
//   - POST /uc/withdraw/apply/code - 申请提现
//   - POST /uc/withdraw/record - 查询提现记录
//   - POST /uc/withdraw/support/coin/info - 查询可提现币种信息
//
// 认证中间件说明：
//   - 使用 serverCtx.Auth 中间件验证 JWT Token
//   - Token 通过 x-auth-token 请求头传递
//   - 中间件会将用户 ID 注入到 context 中
func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 公开路由组：无需 JWT 认证
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/uc/mobile/code", Handler: SendCodeHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/uc/register/phone", Handler: RegisterHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/uc/check/login", Handler: CheckLoginHandler(serverCtx)},
		{Method: http.MethodPost, Path: "/uc/login", Handler: LoginHandler(serverCtx)},
	})

	// 认证路由组：需要 JWT Token 验证
	// 通过 serverCtx.Auth 中间件保护所有路由
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
