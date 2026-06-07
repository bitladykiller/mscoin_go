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
