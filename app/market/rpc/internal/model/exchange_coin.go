package model

// ExchangeCoin 映射 `exchange_coin` 表，捕获面向市场的接口所展示的交易对配置。
//
// 业务用途：
//   - 定义系统支持的交易对
//   - 控制交易对启用/可见状态
//   - 设置交易精度、费率、限制
//   - 配置市场买卖功能
//
// 字段说明：
//   - ID：交易对唯一标识（数据库主键）
//   - Symbol：交易对标识，格式为 "BASEQUOTE"，如 "BTCUSDT"、"ETHUSDT"
//   - BaseSymbol/coinSymbol：基础币种和计价币种单位
//     - 例如 BTCUSDT：BaseSymbol="BTC"（基础币），CoinSymbol="USDT"（计价币）
//   - BaseCoinScale：基础币种精度（价格精度）
//   - CoinScale：计价币种精度（数量精度）
//   - Enable：是否启用交易（1=启用）
//   - Visible：是否对用户可见（1=可见，用于行情展示）
//   - Fee：交易手续费率
//   - Sort：排序权重（用于交易对列表展示顺序）
//   - EnableMarketBuy/EnableMarketSell：是否支持市价买入/卖出（1=支持）
//   - MinSellPrice/MaxBuyPrice：卖出最低价和买入最高价限制
//   - MinVolume/MaxVolume：最小和最大交易量限制
//   - MinTurnover：最小成交额限制
//   - MaxTradingOrder：最大挂单数量
//   - MaxTradingTime：最大挂单时长（秒）
//   - Flag：特殊标记位（如创新区、主板上币等）
//   - Zone：分区标识
//   - PublishType：上币类型（如投票上币、直接上币等）
//   - PublishAmount/PublishPrice：发行量和发行价（用于新币上线）
//   - StartTime/EndTime/ClearTime：上币活动时间配置
//   - Exchangeable：是否可交易（1=可交易）
//   - RobotType：机器人类型（用于做市配置）
type ExchangeCoin struct {
	ID               int64   `db:"id"`
	Symbol           string  `db:"symbol"`
	BaseCoinScale    int64   `db:"base_coin_scale"`
	BaseSymbol       string  `db:"base_symbol"`
	CoinScale        int64   `db:"coin_scale"`
	CoinSymbol       string  `db:"coin_symbol"`
	Enable           int64   `db:"enable"`
	Fee              float64 `db:"fee"`
	Sort             int64   `db:"sort"`
	EnableMarketBuy  int64   `db:"enable_market_buy"`
	EnableMarketSell int64   `db:"enable_market_sell"`
	MinSellPrice     float64 `db:"min_sell_price"`
	Flag             int64   `db:"flag"`
	MaxTradingOrder  int64   `db:"max_trading_order"`
	MaxTradingTime   int64   `db:"max_trading_time"`
	MinTurnover      float64 `db:"min_turnover"`
	ClearTime        int64   `db:"clear_time"`
	EndTime          int64   `db:"end_time"`
	Exchangeable     int64   `db:"exchangeable"`
	MaxBuyPrice      float64 `db:"max_buy_price"`
	MaxVolume        float64 `db:"max_volume"`
	MinVolume        float64 `db:"min_volume"`
	PublishAmount    float64 `db:"publish_amount"`
	PublishPrice     float64 `db:"publish_price"`
	PublishType      int64   `db:"publish_type"`
	RobotType        int64   `db:"robot_type"`
	StartTime        int64   `db:"start_time"`
	Visible          int64   `db:"visible"`
	Zone             int64   `db:"zone"`
}
