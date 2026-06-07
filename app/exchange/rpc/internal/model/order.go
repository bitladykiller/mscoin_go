// Package model 定义了 exchange-rpc 的数据模型。
// 包含订单实体、视图对象以及订单状态、方向、类型的常量定义。
package model

// --- [订单状态] --- //

// 订单状态常量定义
const (
	// OrderTrading 交易中状态，订单正在撮合队列中等待成交。
	OrderTrading = iota
	// OrderCompleted 已完成状态，订单已全部成交。
	OrderCompleted
	// OrderCanceled 已取消状态，订单被用户取消。
	OrderCanceled
	// OrderOverTimed 已超时状态，订单超过有效期被系统取消。
	OrderOverTimed
	// OrderInit 初始化状态，订单刚创建尚未进入撮合队列。
	OrderInit
)

// StatusLabels 是订单状态的字符串标签映射，用于前端展示。
var StatusLabels = map[int]string{
	OrderTrading:   "TRADING",
	OrderCompleted: "COMPLETED",
	OrderCanceled:  "CANCELED",
	OrderOverTimed: "OVERTIMED",
}

// --- [订单方向] --- //

// 订单方向常量定义
const (
	// DirectionBuy 买入方向。
	DirectionBuy = iota
	// DirectionSell 卖出方向。
	DirectionSell
)

// DirectionLabels 是订单方向的字符串标签映射。
var DirectionLabels = map[int]string{
	DirectionBuy:  "BUY",
	DirectionSell: "SELL",
}

// --- [订单类型] --- //

// 订单类型常量定义
const (
	// TypeMarketPrice 市价单，按市场当前最优价格立即成交。
	TypeMarketPrice = iota
	// TypeLimitPrice 限价单，按指定价格挂单等待成交。
	TypeLimitPrice
)

// TypeLabels 是订单类型的字符串标签映射。
var TypeLabels = map[int]string{
	TypeMarketPrice: "MARKET_PRICE",
	TypeLimitPrice:  "LIMIT_PRICE",
}

// directionCodes 是订单方向字符串到代码的映射。
var directionCodes = map[string]int{
	"BUY":  DirectionBuy,
	"SELL": DirectionSell,
}

// typeCodes 是订单类型字符串到代码的映射。
var typeCodes = map[string]int{
	"MARKET_PRICE": TypeMarketPrice,
	"LIMIT_PRICE":  TypeLimitPrice,
}

// ExchangeOrder 对应 exchange_order 数据库表。
// 表示一个交易订单的完整信息。
type ExchangeOrder struct {
	// ID 是数据库自增主键。
	ID int64 `db:"id"`
	// OrderId 是业务订单号，格式为 "E" + 时间戳纳秒。
	OrderId string `db:"order_id"`
	// Amount 是订单总数量。
	Amount float64 `db:"amount"`
	// BaseSymbol 是基础币种符号，如 USDT。
	BaseSymbol string `db:"base_symbol"`
	// CanceledTime 是订单取消时间戳（毫秒）。
	CanceledTime int64 `db:"canceled_time"`
	// CoinSymbol 是交易币种符号，如 BTC。
	CoinSymbol string `db:"coin_symbol"`
	// CompletedTime 是订单完成时间戳（毫秒）。
	CompletedTime int64 `db:"completed_time"`
	// Direction 是订单方向（0:买入, 1:卖出）。
	Direction int `db:"direction"`
	// MemberId 是会员 ID。
	MemberId int64 `db:"member_id"`
	// Price 是订单价格，市价单为 0。
	Price float64 `db:"price"`
	// Status 是订单状态（0:交易中, 1:已完成, 2:已取消, 3:已超时, 4:初始化）。
	Status int `db:"status"`
	// Symbol 是交易对符号，如 "BTCUSDT"。
	Symbol string `db:"symbol"`
	// Time 是订单创建时间戳（毫秒）。
	Time int64 `db:"time"`
	// TradedAmount 是已成交数量。
	TradedAmount float64 `db:"traded_amount"`
	// Turnover 是已成交金额。
	Turnover float64 `db:"turnover"`
	// Type 是订单类型（0:市价, 1:限价）。
	Type int `db:"type"`
	// UseDiscount 是使用的折扣金额。
	UseDiscount string `db:"use_discount"`
}

// OrderView 是订单的视图对象，用于对外传输。
// 与 ExchangeOrder 不同，OrderView 使用字符串标签表示状态、方向和类型，
// 便于前端直接展示，无需再做转换。
type OrderView struct {
	// OrderId 是业务订单号。
	OrderId string `json:"orderId"`
	// Amount 是订单总数量。
	Amount float64 `json:"amount"`
	// BaseSymbol 是基础币种符号。
	BaseSymbol string `json:"baseSymbol"`
	// CanceledTime 是订单取消时间戳（毫秒）。
	CanceledTime int64 `json:"canceledTime"`
	// CoinSymbol 是交易币种符号。
	CoinSymbol string `json:"coinSymbol"`
	// CompletedTime 是订单完成时间戳（毫秒）。
	CompletedTime int64 `json:"completedTime"`
	// Direction 是订单方向标签（"BUY" 或 "SELL"）。
	Direction string `json:"direction"`
	// MemberId 是会员 ID。
	MemberId int64 `json:"memberId"`
	// Price 是订单价格。
	Price float64 `json:"price"`
	// Status 是订单状态标签（"TRADING", "COMPLETED", "CANCELED", "OVERTIMED"）。
	Status string `json:"status"`
	// Symbol 是交易对符号。
	Symbol string `json:"symbol"`
	// Time 是订单创建时间戳（毫秒）。
	Time int64 `json:"time"`
	// TradedAmount 是已成交数量。
	TradedAmount float64 `json:"tradedAmount"`
	// Turnover 是已成交金额。
	Turnover float64 `json:"turnover"`
	// Type 是订单类型标签（"MARKET_PRICE" 或 "LIMIT_PRICE"）。
	Type string `json:"type"`
	// UseDiscount 是使用的折扣金额。
	UseDiscount string `json:"useDiscount"`
}

// DirectionCode 将订单方向标签转换为代码。
func DirectionCode(label string) int {
	return directionCodes[label]
}

// TypeCode 将订单类型标签转换为代码。
func TypeCode(label string) int {
	return typeCodes[label]
}

// ToView 将 ExchangeOrder 转换为 OrderView。
// 将数值型的状态、方向、类型转换为字符串标签，便于前端展示。
func (o *ExchangeOrder) ToView() *OrderView {
	return &OrderView{
		OrderId:       o.OrderId,
		Amount:        o.Amount,
		BaseSymbol:    o.BaseSymbol,
		CanceledTime:  o.CanceledTime,
		CoinSymbol:    o.CoinSymbol,
		CompletedTime: o.CompletedTime,
		Direction:     DirectionLabels[o.Direction],
		MemberId:      o.MemberId,
		Price:         o.Price,
		Status:        StatusLabels[o.Status],
		Symbol:        o.Symbol,
		Time:          o.Time,
		TradedAmount:  o.TradedAmount,
		Turnover:      o.Turnover,
		Type:          TypeLabels[o.Type],
		UseDiscount:   o.UseDiscount,
	}
}
