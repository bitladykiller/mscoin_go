// Package service 提供领域服务实现。
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"mscoin_go/app/jobcenter/internal/model"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/pkg/mq/kafka"
	"mscoin_go/pkg/okxx"

	"google.golang.org/grpc"
)

// visibleExchangeCoinFinder 定义 Market RPC 交易所币种查询接口。
type visibleExchangeCoinFinder interface {
	FindExchangeCoinVisible(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error)
}

// klineWriter 定义 K 线数据写入接口。
type klineWriter interface {
	ReplaceBatch(ctx context.Context, symbol string, period string, list []*model.Kline) error
}

// priceCache 定义价格缓存接口。
type priceCache interface {
	SetCtx(ctx context.Context, key string, value any) error
}

// KlineSyncService 负责市场 K 线数据的异步同步流程。
//
// 核心职责：
//   - 从 OKX API 获取 K 线数据
//   - 存储到 MongoDB
//   - 缓存最新价格到 Redis
//   - 发布最新 K 线到 Kafka（供 WebSocket 订阅）
//
// 为什么由 jobcenter 执行此写入端工作：
//   - K 线同步是周期性的、外部 API 驱动的任务
//   - MongoDB K 线持久化是追加/刷新导向的，非请求范围
//   - market-rpc 应专注于查询行为，jobcenter 负责后台数据获取
//
// 调度策略：
//   - 按配置的周期（1m、5m、1h 等）定时执行
//   - 每次执行获取最近时间窗口的数据
//   - 删除 MongoDB 中的尾部重叠数据后重新插入
type KlineSyncService struct {
	// marketClient Market RPC 客户端，用于查询可见的交易对列表
	marketClient visibleExchangeCoinFinder

	// okxClient OKX API 客户端，用于获取 K 线数据
	okxClient okxx.Client

	// repo K 线数据写入 Repository
	repo klineWriter

	// cache Redis 缓存客户端，用于存储最新价格
	cache priceCache

	// publishers Kafka 生产者映射，用于发布最新 K 线
	publishers map[string]kafka.Producer
}

// NewKlineSyncService 创建 K 线同步服务实例。
//
// 参数：
//   - marketClient: Market RPC 客户端
//   - okxClient: OKX API 客户端
//   - repo: K 线数据 Repository
//   - cache: Redis 缓存客户端
//   - publishers: Kafka 生产者映射
//
// 返回：KlineSyncService 实例
func NewKlineSyncService(
	marketClient visibleExchangeCoinFinder,
	okxClient okxx.Client,
	repo klineWriter,
	cache priceCache,
	publishers map[string]kafka.Producer,
) *KlineSyncService {
	return &KlineSyncService{
		marketClient: marketClient,
		okxClient:    okxClient,
		repo:         repo,
		cache:        cache,
		publishers:   publishers,
	}
}

// SyncPeriod 同步指定周期的所有交易对 K 线数据。
//
// 同步流程：
//  1. 从 Market RPC 获取所有可见交易对
//  2. 遍历每个交易对，调用 syncSymbol 同步数据
//  3. 聚合所有错误，返回联合错误
//
// 错误处理：
//   - 单个交易对同步失败不影响其他交易对
//   - 所有错误聚合后返回，便于问题定位
//
// 参数：
//   - ctx: 上下文，支持超时和取消
//   - period: K 线周期（如 "1m"、"5m"、"1h"）
//   - publishLatest: 是否发布最新 K 线到 Kafka
//   - publishTopic: 发布目标 Topic 名称
//
// 返回：
//   - error: 同步过程中的错误，多个错误会被聚合
func (s *KlineSyncService) SyncPeriod(ctx context.Context, period string, publishLatest bool, publishTopic string) error {
	period = strings.TrimSpace(period)
	if period == "" {
		return fmt.Errorf("kline period is required")
	}
	if s.marketClient == nil {
		return fmt.Errorf("market client is not initialized")
	}
	if s.okxClient == nil {
		return fmt.Errorf("okx client is not initialized")
	}
	if s.repo == nil {
		return fmt.Errorf("kline repository is not initialized")
	}

	pairs, err := s.marketClient.FindExchangeCoinVisible(ctx, &marketpb.MarketReq{})
	if err != nil {
		return err
	}
	if pairs == nil || len(pairs.List) == 0 {
		return nil
	}

	var joinedErr error
	for _, pair := range pairs.List {
		if pair == nil || strings.TrimSpace(pair.Symbol) == "" {
			continue
		}
		if err := s.syncSymbol(ctx, pair.Symbol, period, publishLatest, publishTopic); err != nil {
			joinedErr = errors.Join(joinedErr, fmt.Errorf("sync %s %s: %w", pair.Symbol, period, err))
		}
	}
	return joinedErr
}

// syncSymbol 同步单个交易对的 K 线数据。
//
// 同步流程：
//  1. 调用 OKX API 获取 K 线数据
//  2. 转换为 model.Kline 格式
//  3. 写入 MongoDB（删除尾部重叠数据后插入）
//  4. 如果是 1m 周期：
//     - 缓存最新收盘价到 Redis
//     - 发布最新 K 线到 Kafka（如果配置）
//
// OKX API 说明：
//   - instID 格式：BTC-USDT（使用 "-" 替代 "/"）
//   - bar 参数：K 线周期（如 "1m"、"5m"）
//   - 返回数据按时间倒序排列
//
// 参数：
//   - ctx: 上下文
//   - symbol: 交易对符号（如 "BTC/USDT"）
//   - period: K 线周期
//   - publishLatest: 是否发布到 Kafka
//   - publishTopic: 发布目标 Topic
//
// 返回：
//   - error: 同步过程中的错误
func (s *KlineSyncService) syncSymbol(ctx context.Context, symbol string, period string, publishLatest bool, publishTopic string) error {
	candles, err := s.okxClient.FetchCandles(ctx, strings.ReplaceAll(symbol, "/", "-"), period)
	if err != nil {
		return err
	}
	if len(candles) == 0 {
		return nil
	}

	list := make([]*model.Kline, 0, len(candles))
	for _, candle := range candles {
		item := model.NewKlineFromCandle(period, candle)
		if item != nil {
			list = append(list, item)
		}
	}
	if err := s.repo.ReplaceBatch(ctx, symbol, period, list); err != nil {
		return err
	}

	if period == "1m" && len(list) > 0 {
		latest := list[0]
		if s.cache != nil {
			cacheKey := strings.ReplaceAll(symbol, "/", "::") + "::RATE"
			if err := s.cache.SetCtx(ctx, cacheKey, latest.ClosePrice); err != nil {
				return fmt.Errorf("cache latest price: %w", err)
			}
		}
		if publishLatest && strings.TrimSpace(publishTopic) != "" {
			publisher := s.publishers[publishTopic]
			if publisher == nil {
				return fmt.Errorf("kline publisher for topic %s is not initialized", publishTopic)
			}
			payload, err := json.Marshal(latest)
			if err != nil {
				return fmt.Errorf("marshal latest kline: %w", err)
			}
			if err := publisher.PushWithKey(ctx, symbol, string(payload)); err != nil {
				return fmt.Errorf("publish latest kline: %w", err)
			}
		}
	}

	return nil
}
