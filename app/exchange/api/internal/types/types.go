// Package types 定义了 exchange-api 的请求和响应类型。
// 这些类型用于 HTTP 请求参数绑定和响应序列化。
package types

// ExchangeReq 是交易相关的通用请求结构体。
// 用于下单、查询当前订单和历史订单等接口。
type ExchangeReq struct {
	// IP 是客户端 IP 地址，用于风控和日志记录。
	IP string `json:"ip,optional" form:"ip,optional"`
	// Symbol 是交易对符号，如 "BTCUSDT"。
	Symbol string `json:"symbol,optional" form:"symbol,optional"`
	// PageNo 是分页页码，从 1 开始。
	PageNo int64 `json:"pageNo,optional" form:"pageNo,optional"`
	// PageSize 是每页记录数。
	PageSize int64 `json:"pageSize,optional" form:"pageSize,optional"`
	// Price 是订单价格，限价单必填。
	Price float64 `json:"price,optional" form:"price,optional"`
	// Amount 是订单数量。
	Amount float64 `json:"amount,optional" form:"amount,optional"`
	// Direction 是订单方向，"BUY" 或 "SELL"。
	Direction string `json:"direction,optional" form:"direction,optional"`
	// Type 是订单类型，"MARKET_PRICE"（市价）或 "LIMIT_PRICE"（限价）。
	Type string `json:"type,optional" form:"type,optional"`
	// UseDiscount 是使用的折扣金额。
	UseDiscount float64 `json:"useDiscount,optional" form:"useDiscount,optional"`
}

// OrderValid 验证订单请求参数的有效性。
// 要求 Direction 和 Type 字段非空才有效。
func (r *ExchangeReq) OrderValid() bool {
	return r.Direction != "" && r.Type != ""
}
