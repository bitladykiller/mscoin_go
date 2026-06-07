package model

import (
	"time"

	"mscoin_go/pkg/okxx"
)

// Kline mirrors the MongoDB candle document shape read by `market-rpc`.
//
// The structure intentionally stays aligned with the market read model so
// `jobcenter` can act as the writer and `market-rpc` can stay the reader
// without any translation drift.
type Kline struct {
	Period       string  `bson:"period,omitempty" json:"period"`
	OpenPrice    float64 `bson:"openPrice,omitempty" json:"openPrice"`
	HighestPrice float64 `bson:"highestPrice,omitempty" json:"highestPrice"`
	LowestPrice  float64 `bson:"lowestPrice,omitempty" json:"lowestPrice"`
	ClosePrice   float64 `bson:"closePrice,omitempty" json:"closePrice"`
	Time         int64   `bson:"time,omitempty" json:"time"`
	Count        float64 `bson:"count,omitempty" json:"count"`
	Volume       float64 `bson:"volume,omitempty" json:"volume"`
	Turnover     float64 `bson:"turnover,omitempty" json:"turnover"`
	TimeStr      string  `bson:"timeStr,omitempty" json:"timeStr"`
}

func (k *Kline) TableName(symbol string, period string) string {
	return "exchange_kline_" + symbol + "_" + period
}

func NewKlineFromCandle(period string, candle *okxx.Candle) *Kline {
	if candle == nil {
		return nil
	}

	return &Kline{
		Period:       period,
		OpenPrice:    candle.OpenPrice,
		HighestPrice: candle.HighestPrice,
		LowestPrice:  candle.LowestPrice,
		ClosePrice:   candle.ClosePrice,
		Time:         candle.Time,
		Count:        candle.Count,
		Volume:       candle.Volume,
		Turnover:     candle.Turnover,
		TimeStr:      time.UnixMilli(candle.Time).Format("2006-01-02 15:04:05"),
	}
}
