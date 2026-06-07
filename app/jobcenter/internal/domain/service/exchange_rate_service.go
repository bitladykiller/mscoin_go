package service

import (
	"context"
	"fmt"

	"mscoin_go/pkg/okxx"
)

const (
	usdtCNYRateCacheKey = "USDT::CNY::RATE"
)

type exchangeRateCache interface {
	SetCtx(ctx context.Context, key string, value any) error
}

type exchangeRateFetcher interface {
	FetchExchangeRate(ctx context.Context) (*okxx.ExchangeRate, error)
}

// ExchangeRateSyncService owns the asynchronous USD/CNY synchronization task.
type ExchangeRateSyncService struct {
	cache   exchangeRateCache
	fetcher exchangeRateFetcher
}

func NewExchangeRateSyncService(cache exchangeRateCache, fetcher exchangeRateFetcher) *ExchangeRateSyncService {
	return &ExchangeRateSyncService{
		cache:   cache,
		fetcher: fetcher,
	}
}

func (s *ExchangeRateSyncService) SyncUSDCNY(ctx context.Context) error {
	if s.cache == nil {
		return fmt.Errorf("exchange-rate cache is not initialized")
	}
	if s.fetcher == nil {
		return fmt.Errorf("exchange-rate fetcher is not initialized")
	}

	rate, err := s.fetcher.FetchExchangeRate(ctx)
	if err != nil {
		return err
	}
	if rate == nil || rate.USDCNY <= 0 {
		return fmt.Errorf("exchange-rate payload is invalid")
	}

	if err := s.cache.SetCtx(ctx, usdtCNYRateCacheKey, rate.USDCNY); err != nil {
		return fmt.Errorf("cache usd/cny exchange rate: %w", err)
	}
	return nil
}
