// Package service 提供领域服务实现。
package service

import (
	"context"
	"fmt"

	"mscoin_go/pkg/okxx"
)

const (
	// usdtCNYRateCacheKey 是 USDT/CNY 汇率的缓存键。
	// 格式：USDT::CNY::RATE
	// 存储的值为 float64 类型，表示 1 USDT 等于多少 CNY。
	usdtCNYRateCacheKey = "USDT::CNY::RATE"
)

// exchangeRateCache 定义汇率缓存接口。
type exchangeRateCache interface {
	SetCtx(ctx context.Context, key string, value any) error
}

// exchangeRateFetcher 定义汇率获取接口。
type exchangeRateFetcher interface {
	FetchExchangeRate(ctx context.Context) (*okxx.ExchangeRate, error)
}

// ExchangeRateSyncService 负责 USD/CNY 汇率的异步同步任务。
//
// 核心职责：
//   - 从 OKX API 获取 USDT/CNY 实时汇率
//   - 缓存到 Redis，供其他服务查询
//
// 调度策略：
//   - 定时执行（默认间隔由配置决定）
//   - 执行失败不影响下次调度
//   - 汇率数据更新频率较低，短时间失败可接受
//
// 依赖说明：
//   - cache: Redis 缓存，存储汇率数据
//   - fetcher: OKX API 客户端，获取汇率数据
type ExchangeRateSyncService struct {
	// cache Redis 缓存客户端
	cache exchangeRateCache

	// fetcher OKX API 客户端
	fetcher exchangeRateFetcher
}

// NewExchangeRateSyncService 创建汇率同步服务实例。
//
// 参数：
//   - cache: Redis 缓存客户端
//   - fetcher: OKX API 客户端
//
// 返回：ExchangeRateSyncService 实例
func NewExchangeRateSyncService(cache exchangeRateCache, fetcher exchangeRateFetcher) *ExchangeRateSyncService {
	return &ExchangeRateSyncService{
		cache:   cache,
		fetcher: fetcher,
	}
}

// SyncUSDCNY 同步 USDT/CNY 汇率到缓存。
//
// 同步流程：
//  1. 调用 OKX API 获取汇率数据
//  2. 验证数据有效性（rate.USDCNY > 0）
//  3. 缓存到 Redis
//
// 错误处理：
//   - 缓存未初始化：返回错误
//   - fetcher 未初始化：返回错误
//   - OKX API 调用失败：返回错误，下次调度重试
//   - 汇率数据无效：返回错误
//   - Redis 写入失败：返回错误
//
// 参数：
//   - ctx: 上下文，支持超时和取消
//
// 返回：
//   - error: 同步过程中的错误
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
