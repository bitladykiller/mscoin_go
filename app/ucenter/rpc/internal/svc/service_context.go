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

type ServiceContext struct {
	Config config.Config
	DB     *sqlx.DB
	Cache  *redisx.Client
	Queue  kafka.Producer

	MemberService      *service.MemberService
	WalletService      *service.WalletService
	TransactionService *service.TransactionService
	WithdrawService    *service.WithdrawService
	MarketClient       marketpb.MarketClient
	AddressAllocator   btcx.AddressAllocator
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := mysqlx.New(c.Mysql)
	if err != nil {
		panic(err)
	}
	marketClient := zrpc.MustNewClient(c.MarketRPC)
	cache := redisx.New(c.Redis)
	queue, err := kafka.NewProducer(c.Kafka)
	if err != nil {
		panic(err)
	}
	addressAllocator, err := btcx.NewAddressAllocator(c.Bitcoin)
	if err != nil {
		panic(err)
	}

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

func (s *ServiceContext) Close() {
	if s.Queue != nil {
		_ = s.Queue.Close()
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}
}
