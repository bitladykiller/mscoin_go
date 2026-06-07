// Package service 提供 K 线同步服务的单元测试。
//
// 测试覆盖：
//   - 正常流程：验证 K 线数据的获取、存储、缓存和发布
//   - 错误聚合：验证多个交易对同步失败时的错误聚合
package service

import (
	"context"
	"errors"
	"testing"

	"mscoin_go/app/jobcenter/internal/model"
	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/pkg/mq/kafka"
	"mscoin_go/pkg/okxx"

	"google.golang.org/grpc"
)

// fakeVisibleExchangeCoinFinder 是 visibleExchangeCoinFinder 的 mock 实现。
type fakeVisibleExchangeCoinFinder struct {
	findFn func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error)
}

func (f *fakeVisibleExchangeCoinFinder) FindExchangeCoinVisible(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error) {
	return f.findFn(ctx, in, opts...)
}

// fakeKlineWriter 是 klineWriter 的 mock 实现。
type fakeKlineWriter struct {
	replaceBatchFn func(ctx context.Context, symbol string, period string, list []*model.Kline) error
}

func (f *fakeKlineWriter) ReplaceBatch(ctx context.Context, symbol string, period string, list []*model.Kline) error {
	return f.replaceBatchFn(ctx, symbol, period, list)
}

// fakePriceCache 是 priceCache 的 mock 实现。
type fakePriceCache struct {
	setFn func(ctx context.Context, key string, value any) error
}

func (f *fakePriceCache) SetCtx(ctx context.Context, key string, value any) error {
	return f.setFn(ctx, key, value)
}

// fakeOKXClient 是 okxx.Client 的 mock 实现。
type fakeOKXClient struct {
	fetchExchangeRateFn func(ctx context.Context) (*okxx.ExchangeRate, error)
	fetchCandlesFn      func(ctx context.Context, instID string, bar string) ([]*okxx.Candle, error)
}

func (f *fakeOKXClient) FetchExchangeRate(ctx context.Context) (*okxx.ExchangeRate, error) {
	return f.fetchExchangeRateFn(ctx)
}

func (f *fakeOKXClient) FetchCandles(ctx context.Context, instID string, bar string) ([]*okxx.Candle, error) {
	return f.fetchCandlesFn(ctx, instID, bar)
}

// fakeKlineProducer 是 kafka.Producer 的 mock 实现。
type fakeKlineProducer struct {
	pushFn func(ctx context.Context, key string, value string) error
}

func (f *fakeKlineProducer) PushWithKey(ctx context.Context, key string, value string) error {
	return f.pushFn(ctx, key, value)
}

func (f *fakeKlineProducer) Close() error {
	return nil
}

// TestKlineSyncServiceSyncPeriodStoresAndPublishesLatest1m 验证 1m 周期 K 线的完整同步流程。
//
// 测试场景：
//   - 配置一个交易对（BTC/USDT）
//   - OKX API 返回 K 线数据
//   - 1m 周期，启用发布到 Kafka
//
// 验证点：
//   - OKX API 使用正确的参数（instID=BTC-USDT, bar=1m）
//   - MongoDB ReplaceBatch 被调用
//   - 最新价格缓存到 Redis（key=BTC::USDT::RATE）
//   - 最新 K 线发布到 Kafka（key=BTC/USDT）
func TestKlineSyncServiceSyncPeriodStoresAndPublishesLatest1m(t *testing.T) {
	t.Parallel()

	var (
		cachedKey     string
		cachedValue   any
		publishedKey  string
		replaceCalled bool
	)

	service := NewKlineSyncService(
		&fakeVisibleExchangeCoinFinder{
			findFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error) {
				return &marketpb.ExchangeCoinRes{
					List: []*marketpb.ExchangeCoin{
						{Symbol: "BTC/USDT"},
					},
				}, nil
			},
		},
		&fakeOKXClient{
			fetchCandlesFn: func(ctx context.Context, instID string, bar string) ([]*okxx.Candle, error) {
				if instID != "BTC-USDT" || bar != "1m" {
					t.Fatalf("FetchCandles() args = (%q,%q), want (BTC-USDT,1m)", instID, bar)
				}
				return []*okxx.Candle{
					{
						Time:         1710000000000,
						OpenPrice:    1,
						HighestPrice: 2,
						LowestPrice:  0.5,
						ClosePrice:   1.5,
						Count:        10,
						Volume:       20,
						Turnover:     30,
					},
				}, nil
			},
		},
		&fakeKlineWriter{
			replaceBatchFn: func(ctx context.Context, symbol string, period string, list []*model.Kline) error {
				replaceCalled = true
				if symbol != "BTC/USDT" || period != "1m" {
					t.Fatalf("ReplaceBatch() args = (%q,%q), want (BTC/USDT,1m)", symbol, period)
				}
				if len(list) != 1 || list[0].ClosePrice != 1.5 {
					t.Fatalf("ReplaceBatch() list = %+v, want one latest kline", list)
				}
				return nil
			},
		},
		&fakePriceCache{
			setFn: func(ctx context.Context, key string, value any) error {
				cachedKey = key
				cachedValue = value
				return nil
			},
		},
		map[string]kafka.Producer{
			"kline_1m": &fakeKlineProducer{
				pushFn: func(ctx context.Context, key string, value string) error {
					publishedKey = key
					if value == "" {
						t.Fatal("published payload should not be empty")
					}
					return nil
				},
			},
		},
	)

	if err := service.SyncPeriod(context.Background(), "1m", true, "kline_1m"); err != nil {
		t.Fatalf("SyncPeriod() error = %v", err)
	}
	if !replaceCalled {
		t.Fatal("ReplaceBatch() should be called")
	}
	if cachedKey != "BTC::USDT::RATE" || cachedValue != 1.5 {
		t.Fatalf("cache = (%q,%v), want (BTC::USDT::RATE,1.5)", cachedKey, cachedValue)
	}
	if publishedKey != "BTC/USDT" {
		t.Fatalf("published key = %q, want BTC/USDT", publishedKey)
	}
}

// TestKlineSyncServiceSyncPeriodAggregatesSymbolErrors 验证错误聚合行为。
//
// 测试场景：
//   - OKX API 调用失败
//
// 预期行为：
//   - 返回错误，包含具体的交易对和周期信息
//   - 错误信息便于问题定位
func TestKlineSyncServiceSyncPeriodAggregatesSymbolErrors(t *testing.T) {
	t.Parallel()

	service := NewKlineSyncService(
		&fakeVisibleExchangeCoinFinder{
			findFn: func(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error) {
				return &marketpb.ExchangeCoinRes{
					List: []*marketpb.ExchangeCoin{
						{Symbol: "BTC/USDT"},
					},
				}, nil
			},
		},
		&fakeOKXClient{
			fetchCandlesFn: func(ctx context.Context, instID string, bar string) ([]*okxx.Candle, error) {
				return nil, errors.New("okx unavailable")
			},
		},
		&fakeKlineWriter{
			replaceBatchFn: func(ctx context.Context, symbol string, period string, list []*model.Kline) error {
				return nil
			},
		},
		&fakePriceCache{
			setFn: func(ctx context.Context, key string, value any) error { return nil },
		},
		nil,
	)

	if err := service.SyncPeriod(context.Background(), "1m", false, ""); err == nil {
		t.Fatal("SyncPeriod() should fail when symbol sync fails")
	}
}
