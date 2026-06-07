package service

import (
	"context"
	"time"

	"mscoin_go/app/market/rpc/internal/model"
	"mscoin_go/app/market/rpc/internal/repository"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// MarketService 协调读密集型的市场业务流程，
// 结合交易对元数据和历史 K 线数据。
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

// SymbolThumbTrend 计算每个可见交易对的当前快照列表。
// 该方法有意保留旧项目的行为：当 K 线数据缺失时，
// 回退为空的缩略图而不是让整个请求失败。
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

// HistoryKline 返回市场 RPC 契约所期望的精确传输格式的 K 线数据。
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
