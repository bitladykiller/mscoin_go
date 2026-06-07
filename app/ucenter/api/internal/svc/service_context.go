// Package svc 提供 ucenter-api 服务的上下文管理。
//
// ServiceContext 是服务运行时的依赖容器，持有：
//   - 配置对象
//   - 认证中间件
//   - RPC 客户端连接
//
// 该包负责初始化所有 RPC 客户端连接，并将其注入到各个 logic 层。
package svc

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/api/internal/config"
	"mscoin_go/app/ucenter/api/internal/middleware"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 是 ucenter-api 服务的运行时上下文容器。
//
// 该结构体采用依赖注入模式，集中管理所有服务依赖：
//   - Config：服务配置
//   - Auth：JWT 认证中间件，用于保护需要登录的接口
//   - RegisterClient：注册服务客户端，调用 ucenter-rpc 的注册功能
//   - LoginClient：登录服务客户端，调用 ucenter-rpc 的登录功能
//   - MemberClient：会员服务客户端，调用 ucenter-rpc 的会员管理功能
//   - AssetClient：资产服务客户端，调用 ucenter-rpc 的钱包资产功能
//   - WithdrawClient：提现服务客户端，调用 ucenter-rpc 的提现功能
//   - MarketClient：市场服务客户端，调用 market-rpc 的币种查询功能
//
// RPC 调用关系：
//   - RegisterClient -> ucenter-rpc (register pb) -> 处理用户注册
//   - LoginClient -> ucenter-rpc (login pb) -> 处理用户登录验证
//   - MemberClient -> ucenter-rpc (member pb) -> 查询/更新会员信息
//   - AssetClient -> ucenter-rpc (asset pb) -> 查询钱包余额、交易记录、重置地址
//   - WithdrawClient -> ucenter-rpc (withdraw pb) -> 提现申请、提现记录、地址管理
//   - MarketClient -> market-rpc (market pb) -> 查询币种信息（提现限额、手续费等）
type ServiceContext struct {
	// Config 是服务配置，包含 RPC 连接信息和 JWT 配置。
	Config config.Config

	// Auth 是 JWT 认证中间件，验证请求中的 x-auth-token 头部。
	// 用于保护需要用户登录才能访问的接口。
	Auth rest.Middleware

	// RegisterClient 是用户注册服务的 RPC 客户端。
	// 调用 ucenter-rpc 的 Register 服务，提供手机号注册和验证码发送功能。
	RegisterClient registerpb.RegisterClient

	// LoginClient 是用户登录服务的 RPC 客户端。
	// 调用 ucenter-rpc 的 Login 服务，验证用户名密码并返回 JWT Token。
	LoginClient loginpb.LoginClient

	// MemberClient 是会员管理服务的 RPC 客户端。
	// 调用 ucenter-rpc 的 Member 服务，查询会员详情、安全设置等信息。
	MemberClient memberpb.MemberClient

	// AssetClient 是资产管理服务的 RPC 客户端。
	// 调用 ucenter-rpc 的 Asset 服务，查询钱包余额、交易记录、重置充值地址等。
	AssetClient assetpb.AssetClient

	// WithdrawClient 是提现服务的 RPC 客户端。
	// 调用 ucenter-rpc 的 Withdraw 服务，处理提现申请、查询提现记录、管理提现地址。
	WithdrawClient withdrawpb.WithdrawClient

	// MarketClient 是市场服务的 RPC 客户端。
	// 调用 market-rpc 的 Market 服务，查询币种列表和币种详情（用于提现界面展示）。
	MarketClient marketpb.MarketClient
}

// NewServiceContext 创建并初始化服务上下文。
//
// 初始化流程：
//  1. 建立 ucenter-rpc 客户端连接（用于用户、资产、提现等服务）
//  2. 建立 market-rpc 客户端连接（用于币种信息查询）
//  3. 创建 JWT 认证中间件
//  4. 初始化所有 RPC 客户端存根
//
// 参数：
//   - c：服务配置，包含 RPC 连接地址和认证密钥
//
// 返回：
//   - *ServiceContext：初始化完成的服务上下文
//
// 注意：如果 RPC 连接失败，程序会 panic（MustNewClient）
func NewServiceContext(c config.Config) *ServiceContext {
	ucenterClient := zrpc.MustNewClient(c.UcenterRPC)
	ucenterConn := ucenterClient.Conn()
	marketClient := zrpc.MustNewClient(c.MarketRPC)
	marketConn := marketClient.Conn()

	return &ServiceContext{
		Config:         c,
		Auth:           middleware.NewAuthMiddleware(c.JWT.AccessSecret).Handle,
		RegisterClient: registerpb.NewRegisterClient(ucenterConn),
		LoginClient:    loginpb.NewLoginClient(ucenterConn),
		MemberClient:   memberpb.NewMemberClient(ucenterConn),
		AssetClient:    assetpb.NewAssetClient(ucenterConn),
		WithdrawClient: withdrawpb.NewWithdrawClient(ucenterConn),
		MarketClient:   marketpb.NewMarketClient(marketConn),
	}
}
