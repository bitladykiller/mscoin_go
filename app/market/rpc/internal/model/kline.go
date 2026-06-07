package model

import (
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

// Kline represents the MongoDB document used for time-series market candles.
//
// MongoDB remains a justified choice for this dataset because K-line history is
// append-heavy, query-oriented by symbol/time range, and not part of the core
// transactional state that must remain in MySQL.
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

// TableName derives the historical collection name used by the existing
// project. This behavior must remain stable during migration or history reads
// will silently drift to the wrong dataset.
func (k *Kline) TableName(symbol, period string) string {
	return "exchange_kline_" + symbol + "_" + period
}

// ToCoinThumb converts the latest and earliest candles into the summary shape
// expected by the market RPC layer.
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

// DefaultCoinThumb returns an empty response for pairs that do not currently
// have readable K-line data.
func DefaultCoinThumb(symbol string) *marketpb.CoinThumb {
	return &marketpb.CoinThumb{
		Symbol: symbol,
		Trend:  []float64{},
	}
}
