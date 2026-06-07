// Package service 提供汇率同步服务的单元测试。
//
// 测试覆盖：
//   - 正常流程：验证汇率获取和缓存写入
//   - 数据验证：验证无效汇率被拒绝
//   - 错误传播：验证 API 调用失败的错误传播
package service

import (
	"context"
	"errors"
	"testing"

	"mscoin_go/pkg/okxx"
)

// fakeExchangeRateCache 是 exchangeRateCache 的 mock 实现。
type fakeExchangeRateCache struct {
	setFn func(ctx context.Context, key string, value any) error
}

func (f *fakeExchangeRateCache) SetCtx(ctx context.Context, key string, value any) error {
	return f.setFn(ctx, key, value)
}

// fakeExchangeRateFetcher 是 exchangeRateFetcher 的 mock 实现。
type fakeExchangeRateFetcher struct {
	fetchFn func(ctx context.Context) (*okxx.ExchangeRate, error)
}

func (f *fakeExchangeRateFetcher) FetchExchangeRate(ctx context.Context) (*okxx.ExchangeRate, error) {
	return f.fetchFn(ctx)
}

// TestExchangeRateSyncServiceSyncUSDCNYSavesRate 验证正常同步流程。
//
// 测试场景：
//   - OKX API 返回有效汇率（USDCNY = 7.24）
//   - Redis 缓存可用
//
// 验证点：
//   - 缓存键为 "USDT::CNY::RATE"
//   - 缓存值为 7.24
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

// TestExchangeRateSyncServiceSyncUSDCNYRejectsInvalidRate 验证无效汇率被拒绝。
//
// 测试场景：
//   - OKX API 返回无效汇率（USDCNY = 0）
//
// 预期行为：
//   - 返回错误
//   - 不写入缓存
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

// TestExchangeRateSyncServiceSyncUSDCNYPropagatesFetcherError 验证 API 错误传播。
//
// 测试场景：
//   - OKX API 调用失败
//
// 预期行为：
//   - 返回错误
//   - 不写入缓存
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
