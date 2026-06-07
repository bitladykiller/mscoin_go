// Package service 提供领域服务实现。
//
// 领域服务职责：
//   - 封装核心业务逻辑
//   - 协调多个 Repository 和外部服务
//   - 处理业务错误和异常情况
//
// 设计原则：
//   - 领域服务不依赖消息队列等基础设施
//   - 所有依赖通过接口注入，便于测试
//   - 错误分类明确，支持重试策略
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"mscoin_go/app/jobcenter/internal/model"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	"mscoin_go/pkg/btcx"

	goredis "github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
)

const (
	// withdrawTxCacheKeyPrefix 是提现交易缓存键前缀。
	// 完整键格式：JOBCENTER::WITHDRAW::TX::{recordID}
	withdrawTxCacheKeyPrefix = "JOBCENTER::WITHDRAW::TX::"

	// withdrawTxCacheTTL 是提现交易缓存过期时间（24小时）。
	// 该时间窗口内，即使 MySQL 更新失败，重试时也能从缓存恢复 txid。
	withdrawTxCacheTTL = 24 * time.Hour
)

// withdrawRepository 定义提现记录数据访问接口。
// 接口隔离原则：仅声明领域服务需要的方法。
type withdrawRepository interface {
	FindByID(ctx context.Context, id int64) (*model.WithdrawRecord, error)
	MarkSuccess(ctx context.Context, id int64, txID string, dealTime int64) (bool, error)
}

// marketCoinFinder 定义 Market RPC 币种查询接口。
type marketCoinFinder interface {
	FindCoinById(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error)
}

// assetWalletFinder 定义 Ucenter Asset RPC 钱包查询接口。
type assetWalletFinder interface {
	FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error)
}

// txCache 定义 Redis 缓存接口。
type txCache interface {
	GetCtx(ctx context.Context, key string, value any) error
	SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error
}

// WithdrawTxCacheEntry 是链上交易恢复检查点的缓存数据。
//
// 为什么需要此缓存（恢复机制）：
//   - 广播链上交易和更新 MySQL 不是原子操作
//   - 如果交易已发送但 MySQL 更新失败（如网络问题），下次重试必须：
//     - 复用已知的 txid，而非重复发送资金（防止双重支付）
//     - 使用已记录的 dealTime，保证数据一致性
//   - Redis 为 jobcenter 提供轻量级恢复检查点
//   - 无需在此迁移阶段引入更多数据库变更
//
// 字段说明：
//   - TxID: 链上交易哈希，已广播成功的唯一标识
//   - DealTime: 处理完成时间戳（毫秒）
type WithdrawTxCacheEntry struct {
	TxID     string `json:"txId"`
	DealTime int64  `json:"dealTime"`
}

// NonRetryableError 标记不可重试的毒消息类型错误。
//
// Kafka 消费者处理策略：
//   - 遇到 NonRetryableError 时，发送到死信队列（Dead Letter Queue）
//   - 避免无限重试阻塞消费者
//   - 后续可人工处理死信队列中的消息
//
// 适用场景：
//   - 消息格式错误（反序列化失败）
//   - 业务上不可恢复的错误（如不支持处理的币种）
//   - 状态已变更，无法继续处理（如记录已被人工处理）
type NonRetryableError struct {
	cause error
}

// Error 返回错误消息。
func (e *NonRetryableError) Error() string {
	return e.cause.Error()
}

// Unwrap 返回底层错误，支持 errors.Is 和 errors.As。
func (e *NonRetryableError) Unwrap() error {
	return e.cause
}

// NewNonRetryableError 创建不可重试错误包装。
//
// 参数：
//   - err: 底层错误，nil 时返回 nil
//
// 返回：包装后的 NonRetryableError，或 nil
func NewNonRetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{cause: err}
}

// IsNonRetryable 判断错误是否为不可重试类型。
//
// 使用场景：
//   - Kafka 消费者错误分类
//   - 测试验证错误类型
//
// 参数：
//   - err: 待判断的错误
//
// 返回：true 表示不可重试，应发送到死信队列
func IsNonRetryable(err error) bool {
	var target *NonRetryableError
	return errors.As(err, &target)
}

// WithdrawService 负责提现申请的异步执行工作流。
//
// 核心职责：
//   - 处理来自 ucenter-rpc 的提现事件
//   - 查询币种信息和用户钱包地址
//   - 调用 Bitcoin Core 执行链上转账
//   - 更新提现记录状态
//
// 与 Bitcoin Core 的交互流程：
//  1. 通过 btcx.WithdrawSender 接口调用 Bitcoin Core RPC
//  2. 使用用户的热钱包地址作为发送方
//  3. 使用提现申请中的目标地址作为接收方
//  4. 传入总金额和到账金额（用于计算手续费）
//  5. Bitcoin Core 返回交易哈希（txid）
//
// 错误处理和重试机制：
//   - 参数校验失败：NonRetryableError，不重试
//   - 记录不存在（未提交）：可重试，等待事务提交
//   - 记录状态不支持：NonRetryableError，不重试
//   - 币种不支持：NonRetryableError，不重试
//   - RPC 调用失败：可重试
//   - Bitcoin Core 调用失败：可重试
//   - 缓存写入失败：不影响主流程，记录日志
//   - MySQL 更新失败：可重试，但有缓存检查点保护
//
// 恢复机制：
//   - 交易广播成功后，先将 txid 写入 Redis
//   - 然后 MySQL 更新状态
//   - 如果 MySQL 更新失败，重试时从 Redis 读取 txid
//   - 避免重复广播交易（防止双重支付）
type WithdrawService struct {
	// repo 提现记录数据访问对象
	repo withdrawRepository

	// market Market RPC 客户端，用于查询币种信息
	market marketCoinFinder

	// asset Ucenter Asset RPC 客户端，用于查询用户钱包地址
	asset assetWalletFinder

	// cache Redis 缓存客户端，用于存储恢复检查点
	cache txCache

	// bitcoinSend Bitcoin Core 发送器，用于执行链上转账
	bitcoinSend btcx.WithdrawSender
}

// NewWithdrawService 创建提现服务实例。
//
// 参数：
//   - repo: 提现记录 Repository
//   - market: Market RPC 客户端
//   - asset: Ucenter Asset RPC 客户端
//   - cache: Redis 缓存客户端
//   - bitcoinSend: Bitcoin Core 发送器
//
// 返回：WithdrawService 实例
func NewWithdrawService(
	repo withdrawRepository,
	market marketCoinFinder,
	asset assetWalletFinder,
	cache txCache,
	bitcoinSend btcx.WithdrawSender,
) *WithdrawService {
	return &WithdrawService{
		repo:        repo,
		market:      market,
		asset:       asset,
		cache:       cache,
		bitcoinSend: bitcoinSend,
	}
}

// --- [处理流程] --- //

// ProcessApplied 处理已持久化的提现申请事件。
//
// 完整处理流程：
//  1. 参数校验：验证事件必要字段（ID、MemberId、CoinId、Address）
//  2. 查询记录：从数据库获取提现记录详情
//  3. 状态检查：确保记录处于 Processing 状态
//  4. 恢复检查：尝试从缓存恢复已广播的交易
//  5. 币种查询：获取币种信息（如 BTC）
//  6. 钱包查询：获取用户热钱包地址
//  7. 链上转账：调用 Bitcoin Core 广播交易
//  8. 缓存检查点：将 txid 写入 Redis
//  9. 状态更新：更新 MySQL 记录为成功
//
// 幂等性保证：
//   - 通过状态检查确保不重复处理
//   - 通过缓存检查点确保不重复广播
//   - 通过乐观锁更新确保并发安全
//
// 参数：
//   - ctx: 上下文，支持超时和取消
//   - event: 提现事件，包含记录 ID 和基本信息
//
// 返回：
//   - nil: 处理成功
//   - NonRetryableError: 不可重试错误
//   - 其他错误: 可重试错误
func (s *WithdrawService) ProcessApplied(ctx context.Context, event *model.WithdrawRecordEvent) error {
	if event == nil {
		return NewNonRetryableError(errors.New("withdraw event is required"))
	}
	if event.Id <= 0 {
		return NewNonRetryableError(errors.New("withdraw event id is required"))
	}
	if event.MemberId <= 0 {
		return NewNonRetryableError(errors.New("withdraw event member id is required"))
	}
	if event.CoinId <= 0 {
		return NewNonRetryableError(errors.New("withdraw event coin id is required"))
	}
	if strings.TrimSpace(event.Address) == "" {
		return NewNonRetryableError(errors.New("withdraw event address is required"))
	}

	record, err := s.repo.FindByID(ctx, event.Id)
	if err != nil {
		return err
	}
	if record == nil {
		// 生产者目前在提交外层 SQL 事务之前发布 Kafka 事件。
		// 因此需要短暂的等待窗口，让 jobcenter 在提交可见后能看到该行记录。
		return fmt.Errorf("withdraw record %d is not committed yet", event.Id)
	}
	if record.Status == model.WithdrawStatusSuccess {
		return nil
	}
	if record.Status != model.WithdrawStatusProcessing {
		return NewNonRetryableError(fmt.Errorf("withdraw record %d is in unsupported status %d", record.Id, record.Status))
	}

	if finalized, err := s.finalizeFromCache(ctx, record.Id); err != nil {
		return err
	} else if finalized {
		return nil
	}

	coin, err := s.market.FindCoinById(ctx, &marketpb.MarketReq{Id: record.CoinId})
	if err != nil {
		return err
	}
	if coin == nil || strings.TrimSpace(coin.Unit) == "" {
		return fmt.Errorf("coin %d is unavailable", record.CoinId)
	}
	if coin.Unit != "BTC" {
		return NewNonRetryableError(fmt.Errorf("withdraw coin %s is not implemented in jobcenter yet", coin.Unit))
	}

	wallet, err := s.asset.FindWalletBySymbol(ctx, &assetpb.AssetReq{
		UserId:   record.MemberId,
		CoinName: coin.Unit,
	})
	if err != nil {
		return err
	}
	if wallet == nil || strings.TrimSpace(wallet.Address) == "" {
		return fmt.Errorf("member wallet address is unavailable for user=%d coin=%s", record.MemberId, coin.Unit)
	}

	txID, err := s.bitcoinSend.Send(ctx, wallet.Address, record.Address, record.TotalAmount, record.ArrivedAmount)
	if err != nil {
		return err
	}

	dealTime := time.Now().UnixMilli()
	cacheErr := s.cache.SetWithExpireCtx(ctx, withdrawTxCacheKey(record.Id), WithdrawTxCacheEntry{
		TxID:     txID,
		DealTime: dealTime,
	}, withdrawTxCacheTTL)

	updated, updateErr := s.repo.MarkSuccess(ctx, record.Id, txID, dealTime)
	if updateErr != nil {
		if cacheErr != nil {
			return NewNonRetryableError(fmt.Errorf("chain tx already broadcast but both cache checkpoint and mysql finalization failed: cache=%v mysql=%w", cacheErr, updateErr))
		}
		return updateErr
	}
	if cacheErr != nil {
		return nil
	}
	if !updated {
		return NewNonRetryableError(fmt.Errorf("withdraw record %d status changed before success finalization", record.Id))
	}
	return nil
}

// finalizeFromCache 尝试从缓存恢复已广播的交易并完成状态更新。
//
// 恢复场景：
//   - 交易已广播成功（txid 已获得）
//   - Redis 缓存检查点已写入
//   - MySQL 更新失败（网络问题、服务重启等）
//   - Kafka 消息重试时，从缓存读取 txid
//   - 直接更新 MySQL，无需重新广播交易
//
// 参数：
//   - ctx: 上下文
//   - recordID: 提现记录 ID
//
// 返回：
//   - bool: true 表示已从缓存恢复并完成更新，false 表示缓存无数据
//   - error: 恢复过程中的错误
func (s *WithdrawService) finalizeFromCache(ctx context.Context, recordID int64) (bool, error) {
	var entry WithdrawTxCacheEntry
	if err := s.cache.GetCtx(ctx, withdrawTxCacheKey(recordID), &entry); err != nil {
		if errors.Is(err, goredis.Nil) {
			return false, nil
		}
		return false, fmt.Errorf("load withdraw tx checkpoint: %w", err)
	}
	if strings.TrimSpace(entry.TxID) == "" || entry.DealTime <= 0 {
		return false, nil
	}

	_, err := s.repo.MarkSuccess(ctx, recordID, entry.TxID, entry.DealTime)
	if err != nil {
		return false, err
	}
	return true, nil
}

// withdrawTxCacheKey 生成提现交易缓存键。
//
// 格式：JOBCENTER::WITHDRAW::TX::{recordID}
//
// 参数：
//   - recordID: 提现记录 ID
//
// 返回：Redis 缓存键
func withdrawTxCacheKey(recordID int64) string {
	return fmt.Sprintf("%s%d", withdrawTxCacheKeyPrefix, recordID)
}
