package svc

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/config"
	"mscoin_go/app/ucenter/rpc/internal/domain/service"
	"mscoin_go/app/ucenter/rpc/internal/repository"
	"mscoin_go/pkg/btcx"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/db/mysqlx"
	"mscoin_go/pkg/mq/kafka"

	"github.com/jmoiron/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

// ServiceContext 服务上下文
// 包含 RPC 服务所需的所有依赖项
type ServiceContext struct {
	Config config.Config // 服务配置
	DB     *sqlx.DB      // 数据库连接
	Cache  *redisx.Client // Redis 缓存客户端
	Queue  kafka.Producer // Kafka 消息生产者

	MemberService      *service.MemberService      // 会员服务
	WalletService      *service.WalletService      // 钱包服务
	TransactionService *service.TransactionService // 交易服务
	WithdrawService    *service.WithdrawService    // 提现服务
	MarketClient       marketpb.MarketClient       // Market RPC 客户端
	AddressAllocator   btcx.AddressAllocator       // BTC 地址分配器
}

// NewServiceContext 创建服务上下文实例
// 初始化数据库、缓存、消息队列及所有业务服务
func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	db, err := mysqlx.New(c.Mysql)
	if err != nil {
		panic(err)
	}

	// 初始化 Market RPC 客户端
	marketClient := zrpc.MustNewClient(c.MarketRPC)

	// 初始化 Redis 缓存客户端
	cache := redisx.New(c.Redis)

	// 初始化 Kafka 生产者
	queue, err := kafka.NewProducer(c.Kafka)
	if err != nil {
		panic(err)
	}

	// 初始化 BTC 地址分配器
	addressAllocator, err := btcx.NewAddressAllocator(c.Bitcoin)
	if err != nil {
		panic(err)
	}

	// 初始化各仓储实例
	memberRepo := repository.NewMemberRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	memberAddressRepo := repository.NewMemberAddressRepository(db)
	withdrawRepo := repository.NewWithdrawRepository(db)
	captchaService := service.NewCaptchaService()
	txManager := mysqlx.NewTxManager(db)

	return &ServiceContext{
		Config:             c,
		DB:                 db,
		Cache:              cache,
		Queue:              queue,
		MemberService:      service.NewMemberService(memberRepo, captchaService, cache, c),
		WalletService:      service.NewWalletService(walletRepo),
		TransactionService: service.NewTransactionService(transactionRepo),
		WithdrawService:    service.NewWithdrawService(memberRepo, walletRepo, memberAddressRepo, withdrawRepo, cache, txManager, queue),
		MarketClient:       marketpb.NewMarketClient(marketClient.Conn()),
		AddressAllocator:   addressAllocator,
	}
}

// Close 关闭服务上下文
// 释放数据库连接和消息队列等资源
func (s *ServiceContext) Close() {
	if s.Queue != nil {
		_ = s.Queue.Close()
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}
}
