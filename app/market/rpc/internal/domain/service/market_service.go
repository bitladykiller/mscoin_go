package service

import (
	"context"
	"time"

	"mscoin_go/app/market/rpc/internal/model"
	"mscoin_go/app/market/rpc/internal/repository"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// MarketService coordinates read-heavy market business flows that combine
// trading-pair metadata with historical K-line data.
type MarketService struct {
	klineRepo           *repository.KlineRepository
	exchangeCoinService *ExchangeCoinService
}

func NewMarketService(
	klineRepo *repository.KlineRepository,
	exchangeCoinService *ExchangeCoinService,
) *MarketService {
	return &MarketService{
		klineRepo:           klineRepo,
		exchangeCoinService: exchangeCoinService,
	}
}

// SymbolThumbTrend calculates the current snapshot list for every visible
// trading pair. The method intentionally preserves the old behavior where
// missing K-line data falls back to an empty thumb instead of failing the whole
// request.
func (s *MarketService) SymbolThumbTrend(ctx context.Context) ([]*marketpb.CoinThumb, error) {
	coins, err := s.exchangeCoinService.FindVisible(ctx)
	if err != nil {
		return nil, err
	}

	thumbs := make([]*marketpb.CoinThumb, len(coins))
	from := zeroTimeMillis()
	to := time.Now().UnixMilli()

	for i, coin := range coins {
		klines, err := s.klineRepo.FindBySymbolTime(ctx, coin.Symbol, "1H", from, to, "")
		if err != nil || len(klines) == 0 {
			thumbs[i] = model.DefaultCoinThumb(coin.Symbol)
			continue
		}

		thumbs[i] = buildThumb(coin.Symbol, klines)
	}

	return thumbs, nil
}

// HistoryKline returns candle rows in the exact transport shape expected by the
// market RPC contract.
func (s *MarketService) HistoryKline(
	ctx context.Context,
	symbol string,
	from int64,
	to int64,
	resolution string,
) ([]*marketpb.History, error) {
	period := resolutionToPeriod(resolution)
	klines, err := s.klineRepo.FindBySymbolTime(ctx, symbol, period, from, to, "asc")
	if err != nil {
		return nil, err
	}

	list := make([]*marketpb.History, len(klines))
	for i, item := range klines {
		list[i] = &marketpb.History{
			Time:   item.Time,
			Open:   item.OpenPrice,
			Close:  item.ClosePrice,
			High:   item.HighestPrice,
			Low:    item.LowestPrice,
			Volume: item.Volume,
		}
	}

	return list, nil
}

func buildThumb(symbol string, klines []*model.Kline) *marketpb.CoinThumb {
	last := klines[0]
	first := klines[len(klines)-1]
	thumb := last.ToCoinThumb(symbol, first)

	trend := make([]float64, len(klines))
	high := klines[0].HighestPrice
	low := klines[0].LowestPrice
	volume := 0.0
	turnover := 0.0

	for i := len(klines) - 1; i >= 0; i-- {
		item := klines[i]
		trend[i] = item.ClosePrice
		if item.HighestPrice > high {
			high = item.HighestPrice
		}
		if item.LowestPrice < low {
			low = item.LowestPrice
		}
		volume += item.Volume
		turnover += item.Turnover
	}

	thumb.Trend = trend
	thumb.High = high
	thumb.Low = low
	thumb.Volume = volume
	thumb.Turnover = turnover
	return thumb
}

func resolutionToPeriod(resolution string) string {
	switch resolution {
	case "30":
		return "30m"
	case "15":
		return "15m"
	case "5":
		return "5m"
	case "1":
		return "1m"
	case "1D":
		return "1D"
	case "1W":
		return "1W"
	case "1M":
		return "1M"
	default:
		return "1H"
	}
}

func zeroTimeMillis() int64 {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start.UnixMilli()
}
