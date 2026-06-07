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

type visibleExchangeCoinFinder interface {
	FindExchangeCoinVisible(ctx context.Context, in *marketpb.MarketReq, opts ...grpc.CallOption) (*marketpb.ExchangeCoinRes, error)
}

type klineWriter interface {
	ReplaceBatch(ctx context.Context, symbol string, period string, list []*model.Kline) error
}

type priceCache interface {
	SetCtx(ctx context.Context, key string, value any) error
}

// KlineSyncService owns the asynchronous market-candle synchronization flow.
//
// Why jobcenter performs this write-side work:
//   - K-line synchronization is periodic and external-API-driven
//   - MongoDB candle persistence is append/refresh oriented, not request scoped
//   - `market-rpc` should stay focused on query behavior, while `jobcenter`
//     handles background data acquisition
type KlineSyncService struct {
	marketClient visibleExchangeCoinFinder
	okxClient    okxx.Client
	repo         klineWriter
	cache        priceCache
	publishers   map[string]kafka.Producer
}

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
