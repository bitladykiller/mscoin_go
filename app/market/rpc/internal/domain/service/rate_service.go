package service

import (
	"context"
	"strconv"
	"strings"

	goredis "github.com/go-redis/redis/v8"
)

const (
	usdtCNYRateCacheKey = "USDT::CNY::RATE"
)

type rateCache interface {
	GetCtx(ctx context.Context, key string, value any) error
}

// RateService owns exchange-rate lookup rules.
//
// Why this service keeps fallback values even after Redis integration:
//   - public rate endpoints should degrade gracefully when async sync tasks are
//     temporarily behind
//   - startup order should not make `market-rpc` unavailable
//   - only the dynamically synchronized currencies need cache reads right now
type RateService struct {
	cache     rateCache
	fallbacks map[string]float64
}

func NewRateService(cache rateCache) *RateService {
	return &RateService{
		cache: cache,
		fallbacks: map[string]float64{
			"CNY": 6.95,
			"JPY": 136.23,
		},
	}
}

func (s *RateService) USDRate(ctx context.Context, unit string) float64 {
	normalized := strings.ToUpper(strings.TrimSpace(unit))
	if normalized == "CNY" && s.cache != nil {
		var raw string
		if err := s.cache.GetCtx(ctx, usdtCNYRateCacheKey, &raw); err == nil {
			if value, parseErr := strconv.ParseFloat(strings.TrimSpace(raw), 64); parseErr == nil && value > 0 {
				return value
			}
		} else if err != nil && err != goredis.Nil {
			// Cache errors intentionally fall back so `market-rpc` remains readable
			// even when Redis is temporarily unavailable.
		}
	}
	return s.fallbacks[normalized]
}
