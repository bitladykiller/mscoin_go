package svc

import (
	"context"
	"log"

	"mscoin_go/app/jobcenter/internal/config"
	domainservice "mscoin_go/app/jobcenter/internal/domain/service"
	"mscoin_go/app/jobcenter/internal/repository"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	"mscoin_go/pkg/btcx"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"
	"mscoin_go/pkg/okxx"
	"mscoin_go/pkg/store/mongox"

	"github.com/jmoiron/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext wires the jobcenter dependency graph.
type ServiceContext struct {
	Config config.Config

	UcenterDB *sqlx.DB
	Cache     *redisx.Client
	Mongo     *mongox.Client
	Queue     map[string]kafka.Producer

	MarketClient marketpb.MarketClient
	AssetClient  assetpb.AssetClient

	WithdrawService         *domainservice.WithdrawService
	ExchangeRateSyncService *domainservice.ExchangeRateSyncService
	KlineSyncService        *domainservice.KlineSyncService
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := mysqlx.New(c.UcenterMysql)
	if err != nil {
		panic(err)
	}

	cache := redisx.New(c.Redis)
	marketConn := zrpc.MustNewClient(c.MarketRPC)
	ucenterConn := zrpc.MustNewClient(c.UcenterRPC)
	marketClient := marketpb.NewMarketClient(marketConn.Conn())
	assetClient := assetpb.NewAssetClient(ucenterConn.Conn())

	bitcoinSender, err := btcx.NewWithdrawSender(c.Bitcoin)
	if err != nil {
		panic(err)
	}
	okxClient, err := okxx.NewClient(c.OKX)
	if err != nil {
		panic(err)
	}

	var mongoClient *mongox.Client
	if c.Mongo.URI != "" {
		mongoClient, err = mongox.New(c.Mongo)
		if err != nil {
			log.Fatalf("init mongo: %v", err)
		}
	}

	withdrawRepo := repository.NewWithdrawRepository(db)
	publishers := buildTopicPublishers(c)
	var klineRepo *repository.KlineRepository
	if mongoClient != nil {
		klineRepo = repository.NewKlineRepository(mongoClient.Database())
	}

	return &ServiceContext{
		Config:                  c,
		UcenterDB:               db,
		Cache:                   cache,
		Mongo:                   mongoClient,
		Queue:                   publishers,
		MarketClient:            marketClient,
		AssetClient:             assetClient,
		WithdrawService:         domainservice.NewWithdrawService(withdrawRepo, marketClient, assetClient, cache, bitcoinSender),
		ExchangeRateSyncService: domainservice.NewExchangeRateSyncService(cache, okxClient),
		KlineSyncService:        domainservice.NewKlineSyncService(marketClient, okxClient, klineRepo, cache, publishers),
	}
}

func (s *ServiceContext) Close() {
	for _, producer := range s.Queue {
		if producer != nil {
			_ = producer.Close()
		}
	}
	if s.Mongo != nil {
		_ = s.Mongo.Disconnect(context.Background())
	}
	if s.UcenterDB != nil {
		_ = s.UcenterDB.Close()
	}
}

func buildTopicPublishers(c config.Config) map[string]kafka.Producer {
	topics := make(map[string]struct{})
	for _, item := range c.Tasks.Klines {
		if item.PublishLatest && item.PublishTopic != "" {
			topics[item.PublishTopic] = struct{}{}
		}
	}

	producers := make(map[string]kafka.Producer, len(topics))
	for topic := range topics {
		producer, err := kafka.NewProducer(kafka.Config{
			Brokers:                c.Kafka.Brokers,
			Topic:                  topic,
			Sync:                   true,
			AllowAutoTopicCreation: c.Kafka.AllowAutoTopicCreate,
		})
		if err != nil {
			panic(err)
		}
		producers[topic] = producer
	}
	return producers
}
