// Package model 定义 jobcenter 服务使用的数据模型。
package model

import (
	"time"

	"mscoin_go/pkg/okxx"
)

// Kline 映射 MongoDB K 线文档结构。
//
// K 线（蜡烛图）数据说明：
//   - Period: K 线周期，如 "1m"、"5m"、"1h"、"1d"
//   - OpenPrice: 开盘价
//   - HighestPrice: 最高价
//   - LowestPrice: 最低价
//   - ClosePrice: 收盘价
//   - Time: 时间戳（毫秒）
//   - Count: 成交笔数
//   - Volume: 成交量（基础货币）
//   - Turnover: 成交额（计价货币）
//   - TimeStr: 格式化时间字符串，便于阅读和查询
//
// 存储说明：
//   - 数据存储在 MongoDB 中
//   - 集合名称由 TableName 方法生成
//   - market-rpc 作为读取方，jobcenter 作为写入方
//   - 结构与 market 读取模型保持一致，避免转换偏差
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

// TableName 生成 MongoDB 集合名称。
//
// 命名规则：exchange_kline_{symbol}_{period}
// 例如：exchange_kline_BTC/USDT_1m
//
// 参数：
//   - symbol: 交易对符号，如 "BTC/USDT"
//   - period: K 线周期，如 "1m"
//
// 返回：MongoDB 集合名称
func (k *Kline) TableName(symbol string, period string) string {
	return "exchange_kline_" + symbol + "_" + period
}

// NewKlineFromCandle 从 OKX API 返回的 Candle 数据创建 Kline 实例。
//
// 转换逻辑：
//   - Period 直接使用传入参数
//   - 其他字段从 Candle 结构映射
//   - TimeStr 从时间戳格式化为 "2006-01-02 15:04:05" 格式
//
// 参数：
//   - period: K 线周期
//   - candle: OKX API 返回的蜡烛数据，可能为 nil
//
// 返回：
//   - *Kline: 转换后的 K 线实例，如果 candle 为 nil 则返回 nil
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
