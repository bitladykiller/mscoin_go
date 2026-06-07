// Package svc 提供服务上下文初始化和依赖管理。
//
// 服务上下文层职责：
//   - 聚合所有运行时依赖（数据库连接、缓存、领域服务）
//   - 提供依赖注入容器，供 server 和 logic 层使用
//   - 管理资源生命周期（启动初始化、关闭清理）
//
// 依赖初始化顺序：
//  1. 基础设施层（MySQL、MongoDB、Redis）
//  2. Repository 层（数据访问对象）
//  3. Domain Service 层（业务服务，依赖 Repository）
//
// 这种分层依赖注入确保：
//   - 各层职责清晰，便于测试和替换
//   - 依赖关系显式化，避免隐式全局状态
//   - 生命周期可控，资源正确释放
package svc

import (
	"context"
	"log"
	"time"

	"mscoin_go/app/market/rpc/internal/config"
	"mscoin_go/app/market/rpc/internal/domain/service"
	"mscoin_go/app/market/rpc/internal/repository"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/store/mongox"

	"github.com/jmoiron/sqlx"
)

// ServiceContext 连接 market RPC 服务的所有运行时依赖。
//
// 此处保留详细的依赖图，以便传输层的 logic 文件无需了解基础设施是如何初始化的。
//
// 包含的依赖：
//   - Config：服务配置
//   - DB：MySQL 连接，用于币种和交易对数据
//   - Mongo：MongoDB 连接，用于 K 线历史数据
//   - Cache：Redis 客户端，用于汇率缓存
//   - CoinService：币种领域服务
//   - ExchangeCoinService：交易对领域服务
//   - MarketService：市场数据领域服务
//   - RateService：汇率领域服务
type ServiceContext struct {
	Config config.Config

	DB    *sqlx.DB
	Mongo *mongox.Client
	Cache *redisx.Client

	CoinService         *service.CoinService
	ExchangeCoinService *service.ExchangeCoinService
	MarketService       *service.MarketService
	RateService         *service.RateService
}

// NewServiceContext 初始化所有运行时依赖并返回服务上下文。
//
// 初始化流程：
//  1. 连接 MySQL（币种、交易对数据源）
//  2. 连接 MongoDB（K 线历史数据源）
//  3. 创建 Repository 实例（数据访问层）
//  4. 创建 Redis 缓存客户端
//  5. 创建 Domain Service 实例（业务逻辑层）
//
// 注意：任何依赖初始化失败都会导致服务启动失败（Fatal），
// 这是故意的——服务不应在部分依赖缺失的情况下启动。
func NewServiceContext(c config.Config) *ServiceContext {
	db, err := mysqlx.New(c.Mysql)
	if err != nil {
		log.Fatalf("init mysql: %v", err)
	}

	mongoClient, err := mongox.New(c.Mongo)
	if err != nil {
		log.Fatalf("init mongo: %v", err)
	}

	coinRepo := repository.NewCoinRepository(db)
	exchangeCoinRepo := repository.NewExchangeCoinRepository(db)
	klineRepo := repository.NewKlineRepository(mongoClient.Database())
	cache := redisx.New(c.Redis)

	exchangeCoinService := service.NewExchangeCoinService(exchangeCoinRepo)

	return &ServiceContext{
		Config: c,
		DB:     db,
		Mongo:  mongoClient,
		Cache:  cache,

		CoinService:         service.NewCoinService(coinRepo),
		ExchangeCoinService: exchangeCoinService,
		MarketService:       service.NewMarketService(klineRepo, exchangeCoinService),
		RateService:         service.NewRateService(cache),
	}
}

// Close 释放长生命周期依赖。
//
// 关闭顺序：
//  1. MongoDB 连接（带 5 秒超时）
//  2. MySQL 连接
//
// 注意：Redis 连接由 go-zero 框架管理，此处无需手动关闭。
// 资源释放失败会被静默忽略，因为关闭错误不应阻止服务停止。
func (s *ServiceContext) Close() {
	if s.Mongo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.Mongo.Disconnect(ctx)
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}
}
