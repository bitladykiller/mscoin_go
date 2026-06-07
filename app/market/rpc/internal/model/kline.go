package model

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// Kline 表示用于时序市场 K 线数据的 MongoDB 文档。
//
// MongoDB 对此数据集是一个合理的选择，因为 K 线历史是追加密集型、
// 按币种/时间范围查询导向的，且不属于必须留在 MySQL 中的核心事务状态。
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
// 此行为在迁移过程中必须保持稳定，否则历史读取会静默漂移到错误的数据集。
func (k *Kline) TableName(symbol, period string) string {
	return "exchange_kline_" + symbol + "_" + period
}

// ToCoinThumb 将最新和最早的 K 线转换为市场 RPC 层期望的摘要格式。
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
func DefaultCoinThumb(symbol string) *marketpb.CoinThumb {
	return &marketpb.CoinThumb{
		Symbol: symbol,
		Trend:  []float64{},
	}
}
