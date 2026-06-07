package service

import (
	"context"
	"errors"
	"testing"

	"mscoin_go/pkg/okxx"
)

type fakeExchangeRateCache struct {
	setFn func(ctx context.Context, key string, value any) error
}

func (f *fakeExchangeRateCache) SetCtx(ctx context.Context, key string, value any) error {
	return f.setFn(ctx, key, value)
}

type fakeExchangeRateFetcher struct {
	fetchFn func(ctx context.Context) (*okxx.ExchangeRate, error)
}

func (f *fakeExchangeRateFetcher) FetchExchangeRate(ctx context.Context) (*okxx.ExchangeRate, error) {
	return f.fetchFn(ctx)
}

func TestExchangeRateSyncServiceSyncUSDCNYSavesRate(t *testing.T) {
	t.Parallel()

	var (
		cachedKey   string
		cachedValue any
	)

	service := NewExchangeRateSyncService(
		&fakeExchangeRateCache{
			setFn: func(ctx context.Context, key string, value any) error {
				cachedKey = key
				cachedValue = value
				return nil
			},
		},
		&fakeExchangeRateFetcher{
			fetchFn: func(ctx context.Context) (*okxx.ExchangeRate, error) {
				return &okxx.ExchangeRate{USDCNY: 7.24}, nil
			},
		},
	)

	if err := service.SyncUSDCNY(context.Background()); err != nil {
		t.Fatalf("SyncUSDCNY() error = %v", err)
	}
	if cachedKey != usdtCNYRateCacheKey {
		t.Fatalf("cache key = %q, want %q", cachedKey, usdtCNYRateCacheKey)
	}
	if cachedValue != 7.24 {
		t.Fatalf("cache value = %v, want 7.24", cachedValue)
	}
}

func TestExchangeRateSyncServiceSyncUSDCNYRejectsInvalidRate(t *testing.T) {
	t.Parallel()

	service := NewExchangeRateSyncService(
		&fakeExchangeRateCache{
			setFn: func(ctx context.Context, key string, value any) error {
				t.Fatal("cache should not be written when rate is invalid")
				return nil
			},
		},
		&fakeExchangeRateFetcher{
			fetchFn: func(ctx context.Context) (*okxx.ExchangeRate, error) {
				return &okxx.ExchangeRate{USDCNY: 0}, nil
			},
		},
	)

	if err := service.SyncUSDCNY(context.Background()); err == nil {
		t.Fatal("SyncUSDCNY() should fail for invalid payload")
	}
}

func TestExchangeRateSyncServiceSyncUSDCNYPropagatesFetcherError(t *testing.T) {
	t.Parallel()

	service := NewExchangeRateSyncService(
		&fakeExchangeRateCache{
			setFn: func(ctx context.Context, key string, value any) error {
				t.Fatal("cache should not be written when fetch fails")
				return nil
			},
		},
		&fakeExchangeRateFetcher{
			fetchFn: func(ctx context.Context) (*okxx.ExchangeRate, error) {
				return nil, errors.New("okx unavailable")
			},
		},
	)

	if err := service.SyncUSDCNY(context.Background()); err == nil {
		t.Fatal("SyncUSDCNY() should fail when fetch fails")
	}
}
