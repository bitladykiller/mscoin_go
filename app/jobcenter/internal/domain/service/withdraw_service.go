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
	withdrawTxCacheKeyPrefix = "JOBCENTER::WITHDRAW::TX::"
	withdrawTxCacheTTL       = 24 * time.Hour
)

type withdrawRepository interface {
	FindByID(ctx context.Context, id int64) (*model.WithdrawRecord, error)
	MarkSuccess(ctx context.Context, id int64, txID string, dealTime int64) (bool, error)
}

type marketCoinFinder interface {
	FindCoinById(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.Coin, error)
}

type assetWalletFinder interface {
	FindWalletBySymbol(ctx context.Context, in *assetpb.AssetReq, opts ...grpc.CallOption) (*assetpb.MemberWallet, error)
}

type txCache interface {
	GetCtx(ctx context.Context, key string, value any) error
	SetWithExpireCtx(ctx context.Context, key string, value any, ttl time.Duration) error
}

// WithdrawTxCacheEntry 是在链节点返回 txid 之后、MySQL 完全更新之前写入的恢复数据。
//
// 为什么需要此缓存：
//   - 广播链上交易和更新 MySQL 不是原子操作
//   - 如果交易已发送但 MySQL 更新失败，下次重试必须复用已知 txid，而非重复发送资金
//   - Redis 为 jobcenter 提供轻量级恢复检查点，无需在此迁移阶段引入更多数据库变更
type WithdrawTxCacheEntry struct {
	TxID     string `json:"txId"`
	DealTime int64  `json:"dealTime"`
}

// NonRetryableError 标记一种不可重试的毒消息类型错误。
//
// Kafka 消费者将此错误转换为死信或确认操作，而非无限重试循环。
type NonRetryableError struct {
	cause error
}

func (e *NonRetryableError) Error() string {
	return e.cause.Error()
}

func (e *NonRetryableError) Unwrap() error {
	return e.cause
}

func NewNonRetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &NonRetryableError{cause: err}
}

func IsNonRetryable(err error) bool {
	var target *NonRetryableError
	return errors.As(err, &target)
}

// WithdrawService 负责首个迁移的异步提现执行工作流。
type WithdrawService struct {
	repo        withdrawRepository
	market      marketCoinFinder
	asset       assetWalletFinder
	cache       txCache
	bitcoinSend btcx.WithdrawSender
}

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

// ProcessApplied 处理由 `ucenter-rpc` 发出的已持久化提现事件。
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

func withdrawTxCacheKey(recordID int64) string {
	return fmt.Sprintf("%s%d", withdrawTxCacheKeyPrefix, recordID)
}
