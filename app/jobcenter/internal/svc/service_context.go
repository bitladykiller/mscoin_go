// Package svc 提供 jobcenter 服务的依赖注入容器。
//
// ServiceContext 聚合了所有领域服务、数据访问对象和外部客户端，
// 实现统一的生命周期管理和资源释放。
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

// ServiceContext 是 jobcenter 的依赖注入容器，组装所有服务依赖。
//
// 职责：
//   - 管理数据库连接（MySQL、MongoDB）
//   - 管理缓存客户端（Redis）
//   - 管理 Kafka 生产者
//   - 管理 RPC 客户端（Market、Ucenter）
//   - 聚合领域服务实例
//
// 生命周期：
//   - 在服务启动时由 NewServiceContext 创建
//   - 在服务关闭时由 Close 释放资源
type ServiceContext struct {
	// Config 服务配置，包含所有外部依赖的连接信息
	Config config.Config

	// UcenterDB MySQL 数据库连接，用于提现记录的读写操作
	UcenterDB *sqlx.DB

	// Cache Redis 缓存客户端，用于交易恢复检查点和实时价格缓存
	Cache *redisx.Client

	// Mongo MongoDB 客户端，用于 K 线数据的存储
	Mongo *mongox.Client

	// Queue Kafka 生产者映射，key 为 Topic 名称
	Queue map[string]kafka.Producer

	// MarketClient Market RPC 客户端，用于查询币种信息和交易所交易对
	MarketClient marketpb.MarketClient

	// AssetClient Ucenter Asset RPC 客户端，用于查询用户钱包地址
	AssetClient assetpb.AssetClient

	// WithdrawService 提现处理领域服务，负责提现事件的完整处理流程
	WithdrawService *domainservice.WithdrawService

	// ExchangeRateSyncService 汇率同步领域服务，负责 USD/CNY 汇率的定时同步
	ExchangeRateSyncService *domainservice.ExchangeRateSyncService

	// KlineSyncService K 线同步领域服务，负责市场 K 线数据的定时同步
	KlineSyncService *domainservice.KlineSyncService
}

// NewServiceContext 创建并初始化 ServiceContext。
//
// 初始化顺序：
//  1. 建立 MySQL 连接（Ucenter 数据库）
//  2. 建立 Redis 连接（缓存和恢复检查点）
//  3. 建立 RPC 客户端（Market、Ucenter）
//  4. 创建 Bitcoin Core 发送器（用于提现执行）
//  5. 创建 OKX 客户端（用于汇率和 K 线数据获取）
//  6. 建立 MongoDB 连接（可选，K 线数据存储）
//  7. 创建 Repository 和领域服务实例
//  8. 构建 Kafka 生产者池
//
// 注意：如果任何必需依赖初始化失败，将 panic 终止服务。
// 这是故意的快速失败策略，避免服务在不可用状态下运行。
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

// Close 释放 ServiceContext 持有的所有资源。
//
// 关闭顺序：
//  1. 关闭所有 Kafka 生产者
//  2. 断开 MongoDB 连接
//  3. 关闭 MySQL 连接
//
// 该方法通过 defer 在服务退出时自动调用，确保资源正确释放。
// 错误被忽略，因为服务关闭时通常不需要处理关闭错误。
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

// buildTopicPublishers 根据配置构建 Kafka 生产者池。
//
// 工作原理：
//  1. 遍历所有 K 线任务配置，收集需要发布的 Topic
//  2. 为每个唯一 Topic 创建一个 Kafka 生产者
//  3. 使用同步发送模式确保消息可靠性
//
// 生产者用于将最新 K 线数据推送到消息队列，供前端实时订阅。
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
