package model

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// Kline 表示用于时序市场 K 线数据的 MongoDB 文档。
//
// MongoDB 对此数据集是一个合理的选择，因为：
//   - K 线历史是追加密集型数据（只写入，极少更新）
//   - 按币种/时间范围查询导向的访问模式
//   - 不属于必须留在 MySQL 中的核心事务状态
//   - 天然适合时序数据存储
//
// 集合命名规则：exchange_kline_{symbol}_{period}
//   - symbol：交易对标识，如 "BTCUSDT"
//   - period：K 线周期，如 "1H"、"1D"、"15m"
//
// 字段说明：
//   - Period：K 线周期标识（如 "1H"、"1D"）
//   - OpenPrice：开盘价
//   - HighestPrice：最高价
//   - LowestPrice：最低价
//   - ClosePrice：收盘价
//   - Time：K 线时间（毫秒时间戳，通常是该周期的开始时间）
//   - Count：成交笔数
//   - Volume：成交量（以基础币种计）
//   - Turnover：成交额（以计价币种计）
type Kline struct {
	Period       string  `bson:"period,omitempty"`
	OpenPrice    float64 `bson:"openPrice,omitempty"`
	HighestPrice float64 `bson:"highestPrice,omitempty"`
	LowestPrice  float64 `bson:"lowestPrice,omitempty"`
	ClosePrice   float64 `bson:"closePrice,omitempty"`
	Time         int64   `bson:"time,omitempty"`
	Count        float64 `bson:"count,omitempty"`
	Volume       float64 `bson:"volume,omitempty"`
	Turnover     float64 `bson:"turnover,omitempty"`
}

// TableName 推导现有项目使用的历史集合名称。
//
// 此行为在迁移过程中必须保持稳定，否则历史读取会静默漂移到错误的数据集。
//
// 集合命名规则：exchange_kline_{symbol}_{period}
// 例如：exchange_kline_BTCUSDT_1H
//
// 参数：
//   - symbol：交易对标识
//   - period：K 线周期
//
// 返回：
//   - string：MongoDB 集合名称
func (k *Kline) TableName(symbol, period string) string {
	return "exchange_kline_" + symbol + "_" + period
}

// ToCoinThumb 将最新和最早的 K 线转换为市场 RPC 层期望的摘要格式。
//
// 计算逻辑：
//   - 涨跌 = 最新收盘价 - 最早收盘价
//   - 涨跌幅 = 涨跌 / 最早收盘价 * 100
//   - 其他字段直接取最新 K 线的值
//
// 参数：
//   - symbol：交易对标识
//   - oldest：当日最早的一根 K 线（用于计算涨跌）
//
// 返回：
//   - *marketpb.CoinThumb：缩略图数据（不含趋势线，趋势线由调用方计算）
func (k *Kline) ToCoinThumb(symbol string, oldest *Kline) *marketpb.CoinThumb {
	change := k.ClosePrice - oldest.ClosePrice
	chg := 0.0
	if oldest.ClosePrice != 0 {
		chg = change / oldest.ClosePrice * 100
	}

	return &marketpb.CoinThumb{
		Symbol:      symbol,
		Open:        k.OpenPrice,
		High:        k.HighestPrice,
		Low:         k.LowestPrice,
		Close:       k.ClosePrice,
		Chg:         chg,
		Change:      change,
		UsdRate:     k.ClosePrice,
		BaseUsdRate: 1,
		DateTime:    k.Time,
	}
}

// DefaultCoinThumb 为当前没有可读 K 线数据的交易对返回空响应。
//
// 业务规则：
//   - 当 K 线数据缺失时，返回空的缩略图而不是错误
//   - 确保单个交易对数据问题不影响整体列表展示
//   - 前端可以根据空趋势线做特殊展示
//
// 参数：
//   - symbol：交易对标识
//
// 返回：
//   - *marketpb.CoinThumb：空的缩略图数据（仅包含 symbol 和空趋势线）
func DefaultCoinThumb(symbol string) *marketpb.CoinThumb {
	return &marketpb.CoinThumb{
		Symbol: symbol,
		Trend:  []float64{},
	}
}
