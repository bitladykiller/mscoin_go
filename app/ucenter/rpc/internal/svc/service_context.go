// Package svc 定义服务上下文，聚合所有业务依赖。
//
// ServiceContext 是 ucenter RPC 服务的核心容器，负责初始化和管理：
//   - 数据库连接池（MySQL）
//   - 缓存客户端（Redis）
//   - 消息生产者（Kafka）
//   - 各业务领域服务（会员、钱包、交易、提现）
//   - 外部 RPC 客户端（Market）
//   - Bitcoin 地址分配器
//
// 采用依赖注入模式，所有服务通过 ServiceContext 获取所需依赖，
// 便于测试时 Mock 和生产时替换具体实现。
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
//
// 设计原则：
//   - 单例模式：每个服务实例共享同一个上下文
//   - 延迟初始化：资源在服务启动时一次性初始化
//   - 资源管理：通过 Close 方法统一释放资源
type ServiceContext struct {
	Config config.Config         // 服务配置
	DB     *sqlx.DB              // 数据库连接池，用于所有持久化操作
	Cache  *redisx.Client        // Redis 缓存客户端，用于验证码、会话等临时数据
	Queue  kafka.Producer        // Kafka 消息生产者，用于发布提现等异步事件

	MemberService      *service.MemberService      // 会员服务：处理注册、登录、信息查询
	WalletService      *service.WalletService      // 钱包服务：处理钱包查询、地址管理
	TransactionService *service.TransactionService // 交易服务：处理交易记录查询
	WithdrawService    *service.WithdrawService    // 提现服务：处理提现申请、验证码发送
	MarketClient       marketpb.MarketClient       // Market RPC 客户端：获取币种信息
	AddressAllocator   btcx.AddressAllocator       // BTC 地址分配器：为会员分配 BTC 充值地址
}

// NewServiceContext 创建服务上下文实例
// 初始化数据库、缓存、消息队列及所有业务服务
//
// 初始化顺序：
//  1. 基础设施层：数据库、Redis、Kafka、Bitcoin 节点
//  2. 仓储层：各数据表的 CRUD 操作封装
//  3. 服务层：业务逻辑处理
//
// 注意：任何初始化失败都会 panic，因为服务无法正常运行
func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化数据库连接
	// 使用连接池提高并发性能
	db, err := mysqlx.New(c.Mysql)
	if err != nil {
		panic(err)
	}

	// 初始化 Market RPC 客户端
	// 用于查询币种信息、汇率等市场数据
	marketClient := zrpc.MustNewClient(c.MarketRPC)

	// 初始化 Redis 缓存客户端
	// 用于存储验证码、会话等临时数据
	cache := redisx.New(c.Redis)

	// 初始化 Kafka 生产者
	// 用于发布提现申请等异步事件
	// 下游 jobcenter 消费事件执行实际的链上转账
	queue, err := kafka.NewProducer(c.Kafka)
	if err != nil {
		panic(err)
	}

	// 初始化 BTC 地址分配器
	// 连接 Bitcoin Core 节点，为会员分配充值地址
	addressAllocator, err := btcx.NewAddressAllocator(c.Bitcoin)
	if err != nil {
		panic(err)
	}

	// 初始化各仓储实例
	// 仓储层封装数据库操作，提供领域服务所需的数据访问能力
	memberRepo := repository.NewMemberRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	memberAddressRepo := repository.NewMemberAddressRepository(db)
	withdrawRepo := repository.NewWithdrawRepository(db)
	captchaService := service.NewCaptchaService()
	txManager := mysqlx.NewTxManager(db)

	// 构建服务上下文
	// 各服务通过依赖注入获取所需的仓储和其他依赖
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
//
// 资源释放顺序：
//  1. Kafka 生产者：确保消息发送完成
//  2. 数据库连接：确保事务完成
//
// 注意：Redis 客户端由 go-zero 管理，无需手动关闭
func (s *ServiceContext) Close() {
	if s.Queue != nil {
		_ = s.Queue.Close()
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}
}
