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

// RateService 持有汇率查询规则。
//
// 为什么该服务在集成 Redis 后仍保留回退值：
//   - 公开汇率接口应在异步同步任务暂时滞后时优雅降级
//   - 启动顺序不应导致 market-rpc 不可用
//   - 目前只有动态同步的货币需要读取缓存
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
			// 缓存错误有意回退，以便即使 Redis 暂时不可用，
			// market-rpc 仍然可读。
		}
	}
	return s.fallbacks[normalized]
}
