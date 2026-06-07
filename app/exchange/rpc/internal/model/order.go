package model

// --- [Order Status] --- //

const (
	OrderTrading = iota
	OrderCompleted
	OrderCanceled
	OrderOverTimed
	OrderInit
)

var StatusLabels = map[int]string{
	OrderTrading:   "TRADING",
	OrderCompleted: "COMPLETED",
	OrderCanceled:  "CANCELED",
	OrderOverTimed: "OVERTIMED",
}

// --- [Order Direction] --- //

const (
	DirectionBuy = iota
	DirectionSell
)

var DirectionLabels = map[int]string{
	DirectionBuy:  "BUY",
	DirectionSell: "SELL",
}

// --- [Order Type] --- //

const (
	TypeMarketPrice = iota
	TypeLimitPrice
)

var TypeLabels = map[int]string{
	TypeMarketPrice: "MARKET_PRICE",
	TypeLimitPrice:  "LIMIT_PRICE",
}

var directionCodes = map[string]int{
	"BUY":  DirectionBuy,
	"SELL": DirectionSell,
}

var typeCodes = map[string]int{
	"MARKET_PRICE": TypeMarketPrice,
	"LIMIT_PRICE":  TypeLimitPrice,
}

// ExchangeOrder mirrors the `exchange_order` table.
type ExchangeOrder struct {
	ID            int64   `db:"id" gorm:"column:id"`
	OrderId       string  `db:"order_id" gorm:"column:order_id"`
	Amount        float64 `db:"amount" gorm:"column:amount"`
	BaseSymbol    string  `db:"base_symbol" gorm:"column:base_symbol"`
	CanceledTime  int64   `db:"canceled_time" gorm:"column:canceled_time"`
	CoinSymbol    string  `db:"coin_symbol" gorm:"column:coin_symbol"`
	CompletedTime int64   `db:"completed_time" gorm:"column:completed_time"`
	Direction     int     `db:"direction" gorm:"column:direction"`
	MemberId      int64   `db:"member_id" gorm:"column:member_id"`
	Price         float64 `db:"price" gorm:"column:price"`
	Status        int     `db:"status" gorm:"column:status"`
	Symbol        string  `db:"symbol" gorm:"column:symbol"`
	Time          int64   `db:"time" gorm:"column:time"`
	TradedAmount  float64 `db:"traded_amount" gorm:"column:traded_amount"`
	Turnover      float64 `db:"turnover" gorm:"column:turnover"`
	Type          int     `db:"type" gorm:"column:type"`
	UseDiscount   string  `db:"use_discount" gorm:"column:use_discount"`
}

// OrderView is the transport-facing version of an order. The old API exposes
// string labels for status, direction and type; this dedicated view keeps that
// mapping out of the repository layer.
type OrderView struct {
	OrderId       string  `json:"orderId"`
	Amount        float64 `json:"amount"`
	BaseSymbol    string  `json:"baseSymbol"`
	CanceledTime  int64   `json:"canceledTime"`
	CoinSymbol    string  `json:"coinSymbol"`
	CompletedTime int64   `json:"completedTime"`
	Direction     string  `json:"direction"`
	MemberId      int64   `json:"memberId"`
	Price         float64 `json:"price"`
	Status        string  `json:"status"`
	Symbol        string  `json:"symbol"`
	Time          int64   `json:"time"`
	TradedAmount  float64 `json:"tradedAmount"`
	Turnover      float64 `json:"turnover"`
	Type          string  `json:"type"`
	UseDiscount   string  `json:"useDiscount"`
}

func DirectionCode(label string) int {
	return directionCodes[label]
}

func TypeCode(label string) int {
	return typeCodes[label]
}

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
