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

// WithdrawTxCacheEntry is the recovery payload written after the chain node
// returns a txid but before MySQL is fully updated.
//
// Why this cache exists:
//   - broadcasting the on-chain transaction and updating MySQL are not one
//     atomic operation
//   - if MySQL update fails after the transaction has been sent, the next retry
//     must reuse the known txid instead of sending funds twice
//   - Redis gives jobcenter a lightweight recovery checkpoint without dragging
//     more schema changes into this migration stage
type WithdrawTxCacheEntry struct {
	TxID     string `json:"txId"`
	DealTime int64  `json:"dealTime"`
}

// NonRetryableError marks one poison-message style failure.
//
// The Kafka consumer turns this error into a dead-letter or acknowledge action
// instead of an infinite retry loop.
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

// WithdrawService owns the first migrated async withdraw execution workflow.
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

// --- [Processing Flow] --- //

// ProcessApplied handles one persisted withdraw event emitted by `ucenter-rpc`.
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
		// The producer currently publishes the Kafka event before committing the
		// surrounding SQL transaction. A short retry window is therefore required
		// so jobcenter can see the row once the commit becomes visible.
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
