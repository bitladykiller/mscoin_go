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
//
// 业务职责：
//   - 聚合交易对数据和 K 线数据生成市场概览
//   - 提供历史 K 线查询服务
//   - 计算价格趋势和涨跌幅
//
// 依赖：
//   - klineRepo：KlineRepository，用于 K 线数据访问（MongoDB）
//   - exchangeCoinService：ExchangeCoinService，用于交易对查询
//
// 调用关系：
//
//	Logic -> MarketService -> KlineRepository
//	                      \-> ExchangeCoinService
type MarketService struct {
	klineRepo           *repository.KlineRepository
	exchangeCoinService *ExchangeCoinService
}

// NewMarketService 创建 MarketService 实例。
//
// 参数：
//   - klineRepo：K 线数据仓库
//   - exchangeCoinService：交易对领域服务
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
//
// 业务规则：
//   - 遍历所有可见交易对
//   - 为每个交易对查询当日 K 线数据
//   - 计算开盘价、最高价、最低价、收盘价、涨跌幅、趋势线
//   - 当 K 线数据缺失时，回退为空的缩略图而不是让整个请求失败
//
// 为什么回退而不是失败：
//   - 市场概览应该尽可能展示数据
//   - 单个交易对数据问题不应影响整体列表
//   - 前端可以根据空数据做特殊展示
//
// 参数：
//   - ctx：请求上下文
//
// 返回：
//   - []*marketpb.CoinThumb：交易对缩略图列表
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
//
// 业务规则：
//   - 将 resolution 转换为内部 period 格式
//   - 按时间范围查询 K 线数据
//   - 转换为 protobuf 格式返回
//
// resolution 到 period 的映射：
//   - "1" -> "1m"（1 分钟）
//   - "5" -> "5m"（5 分钟）
//   - "15" -> "15m"（15 分钟）
//   - "30" -> "30m"（30 分钟）
//   - "1D" -> "1D"（1 天）
//   - "1W" -> "1W"（1 周）
//   - "1M" -> "1M"（1 月）
//   - 其他 -> "1H"（默认 1 小时）
//
// 参数：
//   - ctx：请求上下文
//   - symbol：交易对标识
//   - from：开始时间（毫秒时间戳）
//   - to：结束时间（毫秒时间戳）
//   - resolution：K 线周期
//
// 返回：
//   - []*marketpb.History：K 线历史数据列表
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

// buildThumb 从 K 线数据构建交易对缩略图。
//
// 计算逻辑：
//   - 取最新一根 K 线作为基准
//   - 计算涨跌：最新收盘价 - 当日第一根收盘价
//   - 计算涨跌幅：涨跌 / 第一根收盘价 * 100
//   - 遍历所有 K 线计算当日最高、最低、总成交量、总成交额
//   - 收集收盘价作为趋势线数据
//
// 参数：
//   - symbol：交易对标识
//   - klines：K 线数据（已按时间降序排列）
//
// 返回：
//   - *marketpb.CoinThumb：完整的缩略图数据
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

// resolutionToPeriod 将前端 resolution 参数转换为 MongoDB 集合名中的 period。
//
// 前端使用 TradingView 风格的 resolution，后端存储使用特定 period 格式。
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

// zeroTimeMillis 返回当天零点的毫秒时间戳。
//
// 用于查询当日 K 线数据的起始时间。
func zeroTimeMillis() int64 {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return start.UnixMilli()
}
