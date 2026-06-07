package service

import (
	"context"
	"errors"
	"testing"

	goredis "github.com/go-redis/redis/v8"
)

type fakeRateCache struct {
	getFn func(ctx context.Context, key string, value any) error
}

func (f *fakeRateCache) GetCtx(ctx context.Context, key string, value any) error {
	return f.getFn(ctx, key, value)
}

func TestRateServiceReturnsRedisCNYRate(t *testing.T) {
	t.Parallel()

	service := NewRateService(&fakeRateCache{
		getFn: func(ctx context.Context, key string, value any) error {
			if key != usdtCNYRateCacheKey {
				t.Fatalf("cache key = %q, want %q", key, usdtCNYRateCacheKey)
			}
			target := value.(*string)
			*target = "7.321"
			return nil
		},
	})

	if got := service.USDRate(context.Background(), "cny"); got != 7.321 {
		t.Fatalf("USDRate() = %v, want 7.321", got)
	}
}

func TestRateServiceFallsBackWhenRedisMisses(t *testing.T) {
	t.Parallel()

	service := NewRateService(&fakeRateCache{
		getFn: func(ctx context.Context, key string, value any) error {
			return goredis.Nil
		},
	})

	if got := service.USDRate(context.Background(), "CNY"); got != 6.95 {
		t.Fatalf("USDRate() = %v, want fallback 6.95", got)
	}
}

func TestRateServiceFallsBackWhenRedisFails(t *testing.T) {
	t.Parallel()

	service := NewRateService(&fakeRateCache{
		getFn: func(ctx context.Context, key string, value any) error {
			return errors.New("redis down")
		},
	})

	if got := service.USDRate(context.Background(), "CNY"); got != 6.95 {
		t.Fatalf("USDRate() = %v, want fallback 6.95", got)
	}
}
