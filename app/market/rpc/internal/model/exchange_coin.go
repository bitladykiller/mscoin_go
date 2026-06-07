package model

// ExchangeCoin mirrors the `exchange_coin` table and captures the trading pair
// configuration shown by market-facing endpoints.
type ExchangeCoin struct {
	ID               int64   `db:"id" gorm:"column:id"`
	Symbol           string  `db:"symbol" gorm:"column:symbol"`
	BaseCoinScale    int64   `db:"base_coin_scale" gorm:"column:base_coin_scale"`
	BaseSymbol       string  `db:"base_symbol" gorm:"column:base_symbol"`
	CoinScale        int64   `db:"coin_scale" gorm:"column:coin_scale"`
	CoinSymbol       string  `db:"coin_symbol" gorm:"column:coin_symbol"`
	Enable           int64   `db:"enable" gorm:"column:enable"`
	Fee              float64 `db:"fee" gorm:"column:fee"`
	Sort             int64   `db:"sort" gorm:"column:sort"`
	EnableMarketBuy  int64   `db:"enable_market_buy" gorm:"column:enable_market_buy"`
	EnableMarketSell int64   `db:"enable_market_sell" gorm:"column:enable_market_sell"`
	MinSellPrice     float64 `db:"min_sell_price" gorm:"column:min_sell_price"`
	Flag             int64   `db:"flag" gorm:"column:flag"`
	MaxTradingOrder  int64   `db:"max_trading_order" gorm:"column:max_trading_order"`
	MaxTradingTime   int64   `db:"max_trading_time" gorm:"column:max_trading_time"`
	MinTurnover      float64 `db:"min_turnover" gorm:"column:min_turnover"`
	ClearTime        int64   `db:"clear_time" gorm:"column:clear_time"`
	EndTime          int64   `db:"end_time" gorm:"column:end_time"`
	Exchangeable     int64   `db:"exchangeable" gorm:"column:exchangeable"`
	MaxBuyPrice      float64 `db:"max_buy_price" gorm:"column:max_buy_price"`
	MaxVolume        float64 `db:"max_volume" gorm:"column:max_volume"`
	MinVolume        float64 `db:"min_volume" gorm:"column:min_volume"`
	PublishAmount    float64 `db:"publish_amount" gorm:"column:publish_amount"`
	PublishPrice     float64 `db:"publish_price" gorm:"column:publish_price"`
	PublishType      int64   `db:"publish_type" gorm:"column:publish_type"`
	RobotType        int64   `db:"robot_type" gorm:"column:robot_type"`
	StartTime        int64   `db:"start_time" gorm:"column:start_time"`
	Visible          int64   `db:"visible" gorm:"column:visible"`
	Zone             int64   `db:"zone" gorm:"column:zone"`
}
