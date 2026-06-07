// Package svc 定义了 exchange-rpc 的服务上下文。
// ServiceContext 聚合了所有运行时依赖，包括数据库、缓存、RPC 客户端和领域服务。
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
	MemberClient memberpb.MemberClient
	// AssetClient 是用户中心 RPC 服务的资产客户端。
	AssetClient assetpb.AssetClient
	// MarketClient 是行情 RPC 服务的市场客户端。
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
