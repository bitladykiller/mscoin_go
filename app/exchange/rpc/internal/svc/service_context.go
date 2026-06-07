// Package svc 定义了 exchange-rpc 的服务上下文。
// ServiceContext 聚合了所有运行时依赖，包括数据库、缓存、RPC 客户端和领域服务。
//
// 依赖注入说明：
// - ServiceContext 作为依赖容器，管理所有服务依赖的生命周期
// - Logic 层通过 ServiceContext 访问数据库、缓存和外部 RPC 服务
// - 这种设计使得依赖关系清晰，便于测试和替换实现
//
// 与其他服务的依赖关系：
// - ucenter-rpc: 提供会员信息查询、钱包信息查询
// - market-rpc: 提供交易对配置查询
package svc

import (
	"mscoin_go/app/exchange/rpc/internal/config"
	"mscoin_go/app/exchange/rpc/internal/domain/service"
	"mscoin_go/app/exchange/rpc/internal/repository"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"

	"github.com/jmoiron/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 聚合 exchange-rpc 服务的所有运行时依赖。
//
// 依赖类型：
// 1. 配置：Config - 服务配置信息
// 2. 基础设施：DB（MySQL）、Cache（Redis）
// 3. 领域服务：OrderService - 订单业务逻辑
// 4. RPC 客户端：
//    - MemberClient: ucenter-rpc 会员服务客户端
//    - AssetClient: ucenter-rpc 资产服务客户端
//    - MarketClient: market-rpc 行情服务客户端
//
// 与 ucenter-rpc 的调用关系：
// - MemberClient.FindMemberById(): 查询会员信息，验证用户交易状态
// - AssetClient.FindWalletBySymbol(): 查询钱包信息，验证钱包锁定状态
//
// 与 market-rpc 的调用关系：
// - MarketClient.FindSymbolInfo(): 查询交易对配置，验证交易对状态和价格限制
type ServiceContext struct {
	// Config 是服务配置。
	Config config.Config

	// DB 是 MySQL 数据库连接池。
	DB *sqlx.DB
	// Cache 是 Redis 缓存客户端。
	Cache *redisx.Client

	// OrderService 是订单领域服务，封装订单业务逻辑。
	OrderService *service.OrderService

	// MemberClient 是用户中心 RPC 服务的会员客户端。
	// 用于查询会员基本信息和交易状态。
	// 调用方法：FindMemberById
	MemberClient memberpb.MemberClient
	// AssetClient 是用户中心 RPC 服务的资产客户端。
	// 用于查询用户钱包信息，包括余额和锁定状态。
	// 调用方法：FindWalletBySymbol
	AssetClient assetpb.AssetClient
	// MarketClient 是行情 RPC 服务的市场客户端。
	// 用于查询交易对配置信息，包括价格限制、交易状态等。
	// 调用方法：FindSymbolInfo
	MarketClient marketpb.MarketClient
}

// NewServiceContext 创建 ServiceContext 实例。
// 初始化流程：
// 1. 创建 MySQL 数据库连接
// 2. 创建 ucenter-rpc 和 market-rpc 客户端连接
// 3. 创建订单仓库和领域服务
func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 MySQL 数据库连接
	db, err := mysqlx.New(c.Mysql)
	if err != nil {
		panic(err)
	}

	// 创建 RPC 客户端连接
	ucClient := zrpc.MustNewClient(c.UcenterRPC)
	marketClient := zrpc.MustNewClient(c.MarketRPC)

	// 创建订单仓库
	orderRepo := repository.NewOrderRepository(db)

	return &ServiceContext{
		Config:       c,
		DB:           db,
		Cache:        redisx.New(c.Redis),
		OrderService: service.NewOrderService(orderRepo),
		MemberClient: memberpb.NewMemberClient(ucClient.Conn()),
		AssetClient:  assetpb.NewAssetClient(ucClient.Conn()),
		MarketClient: marketpb.NewMarketClient(marketClient.Conn()),
	}
}

// Close 释放服务上下文持有的资源。
// 关闭数据库连接。
func (s *ServiceContext) Close() {
	if s.DB != nil {
		_ = s.DB.Close()
	}
}
