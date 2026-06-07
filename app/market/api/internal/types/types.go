// Package types 定义 market-api 服务的请求和响应数据结构。
// 这些类型定义了 HTTP 接口的输入输出契约，
// 确保 API 层与前端之间的数据交互格式一致。
//
// 类型命名规范：
//   - *Request: 请求参数结构体
//   - *Response: 响应数据结构体
//   - *Resp: 简化命名的响应结构体
//
// 注意：字段标签使用 go-zero 的路由解析语法，
// 支持 path（路径参数）、form（表单参数）、json（JSON 参数）等绑定方式。
package types

// RateRequest 是法币汇率查询的请求参数结构。
// 用于查询指定法币单位对 USD 的实时汇率。
//
// HTTP 路由: POST /market/exchange-rate/usd/:unit
//
// 示例请求:
//
//	POST /market/exchange-rate/usd/CNY
//	{"ip": "192.168.1.1"}
type RateRequest struct {
	// Unit 法币单位代码，如 "CNY"、"EUR"、"JPY" 等
	// 通过 URL 路径参数传递
	Unit string `path:"unit" json:"unit"`

	// IP 客户端 IP 地址，可选参数
	// 用于基于地理位置的汇率查询优化
	// 通常由 handler 自动填充客户端真实 IP
	IP string `json:"ip,optional"`
}

// RateResponse 是汇率查询的响应结构。
// 返回指定法币对 USD 的汇率值。
//
// 响应示例:
//
//	{"rate": 7.24}
type RateResponse struct {
	// Rate 汇率值，表示 1 USD 可兑换的法币数量
	// 例如 CNY 的 rate 为 7.24 表示 1 USD = 7.24 CNY
	Rate float64 `json:"rate"`
}

// MarketReq 是市场数据查询的通用请求参数结构。
// 该结构被多个市场相关接口复用，字段名与旧 API 保持一致以确保前端兼容。
//
// 适用接口:
//   - POST /market/coin-info: 查询币种信息
//   - GET /market/history: 查询 K 线历史
//   - POST /market/symbol-info: 查询交易对信息
//   - POST /market/symbol-thumb: 查询行情缩略图
//   - POST /market/symbol-thumb-trend: 查询行情趋势
//
// 字段说明:
//   - 大部分字段为可选，根据具体接口需求使用
//   - IP 字段通常由 handler 自动填充
type MarketReq struct {
	// IP 客户端 IP 地址，可选
	// 用于地理位置相关的数据筛选和风控
	IP string `json:"ip,optional" form:"ip,optional"`

	// Symbol 交易对代码，如 "BTCUSDT"，可选
	// 用于查询特定交易对的信息
	Symbol string `json:"symbol,optional" form:"symbol,optional" path:"symbol,optional"`

	// Unit 计价单位/法币单位，可选
	// 用于指定返回数据的计价货币
	Unit string `json:"unit,optional" form:"unit,optional"`

	// From K线查询起始时间戳（毫秒），可选
	// 用于历史 K 线数据的时间范围筛选
	From int64 `json:"from,optional" form:"from,optional"`

	// To K线查询结束时间戳（毫秒），可选
	// 用于历史 K 线数据的时间范围筛选
	To int64 `json:"to,optional" form:"to,optional"`

	// Resolution K线周期，可选
	// 常见值: "1" (1分钟), "5" (5分钟), "15" (15分钟),
	//        "30" (30分钟), "60" (1小时), "1D" (1天) 等
	Resolution string `json:"resolution,optional" form:"resolution,optional"`
}

// CoinThumbResp 是返回给 Web 客户端的市场行情快照模型。
// 包含交易对的实时行情数据，用于首页展示和行情列表。
//
// 数据来源：
//   - 由 market-rpc 服务聚合计算后返回
//   - 数据定期更新，非实时推送
//
// 响应示例:
//
//	{
//	  "symbol": "BTCUSDT",
//	  "open": 42000.0,
//	  "high": 43000.0,
//	  "low": 41500.0,
//	  "close": 42500.0,
//	  "chg": 1.19,
//	  "volume": 123456.78,
//	  "usdRate": 1.0
//	}
type CoinThumbResp struct {
	// Symbol 交易对代码，格式为 "基础货币+计价货币"
	// 例如 "BTCUSDT" 表示 BTC/USDT 交易对
	Symbol string `json:"symbol"`

	// Open 开盘价，24小时周期内的第一笔成交价
	Open float64 `json:"open"`

	// High 最高价，24小时周期内的最高成交价
	High float64 `json:"high"`

	// Low 最低价，24小时周期内的最低成交价
	Low float64 `json:"low"`

	// Close 最新价/收盘价，最近一笔成交价
	Close float64 `json:"close"`

	// Chg 涨跌幅百分比，计算公式: (close - lastDayClose) / lastDayClose * 100
	Chg float64 `json:"chg"`

	// Change 涨跌额，计算公式: close - lastDayClose
	Change float64 `json:"change"`

	// Volume 成交量，24小时周期内的成交数量（基础货币计）
	Volume float64 `json:"volume"`

	// Turnover 成交额，24小时周期内的成交金额（计价货币计）
	Turnover float64 `json:"turnover"`

	// LastDayClose 昨日收盘价，用于计算涨跌幅
	LastDayClose float64 `json:"lastDayClose"`

	// USDTRate 对 USDT 的汇率
	// 如果计价货币是 USDT，则值为 1.0
	USDTRate float64 `json:"usdRate"`

	// BaseUSDTRate 基础货币对 USDT 的汇率
	// 用于跨币种价值换算
	BaseUSDTRate float64 `json:"baseUsdRate"`

	// Zone 交易区域/板块标识
	// 用于区分主版、创新版等不同交易区域
	Zone int `json:"zone"`

	// Trend 价格趋势数据，可选字段
	// 包含一段时间内的价格走势点，用于绘制迷你趋势图
	Trend []float64 `json:"trend,optional"`
}

// ExchangeCoinResp 是交易对元数据的公开模型。
// 包含交易对的完整配置信息，用于交易对详情展示和交易规则说明。
//
// 数据来源：
//   - 从数据库 exchange_coin 表读取
//   - 包含交易规则、限制、状态等配置信息
//
// 使用场景：
//   - 交易对详情页
//   - 交易规则展示
//   - 交易限制提示
type ExchangeCoinResp struct {
	// ID 交易对唯一标识
	ID int64 `json:"id"`

	// Symbol 交易对代码，如 "BTCUSDT"
	Symbol string `json:"symbol"`

	// BaseCoinScale 基础货币精度（小数位数）
	// 如 BTC 精度为 8，表示最多交易 0.00000001 BTC
	BaseCoinScale int64 `json:"baseCoinScale"`

	// BaseSymbol 基础货币代码，如 "BTC"
	BaseSymbol string `json:"baseSymbol"`

	// CoinScale 计价货币精度（小数位数）
	// 如 USDT 精度为 6
	CoinScale int64 `json:"coinScale"`

	// CoinSymbol 计价货币代码，如 "USDT"
	CoinSymbol string `json:"coinSymbol"`

	// Enable 交易对启用状态
	// 1: 启用, 0: 禁用
	Enable int64 `json:"enable"`

	// Fee 交易手续费率
	// 如 0.001 表示 0.1% 手续费
	Fee float64 `json:"fee"`

	// Sort 排序权重，数值越小越靠前
	Sort int64 `json:"sort"`

	// EnableMarketBuy 是否允许市价买入
	// 1: 允许, 0: 禁止
	EnableMarketBuy int64 `json:"enableMarketBuy"`

	// EnableMarketSell 是否允许市价卖出
	// 1: 允许, 0: 禁止
	EnableMarketSell int64 `json:"enableMarketSell"`

	// MinSellPrice 最小卖出价格限制
	MinSellPrice float64 `json:"minSellPrice"`

	// Flag 交易对标记/状态位
	// 用于存储额外的状态标识
	Flag int64 `json:"flag"`

	// MaxTradingOrder 最大挂单数量限制
	MaxTradingOrder int64 `json:"maxTradingOrder"`

	// MaxTradingTime 最大持仓时间限制（秒）
	MaxTradingTime int64 `json:"maxTradingTime"`

	// MinTurnover 最小成交额限制
	// 单笔交易金额必须大于此值
	MinTurnover float64 `json:"minTurnover"`

	// ClearTime 清算时间设置
	ClearTime int64 `json:"clearTime"`

	// EndTime 交易结束时间（时间戳）
	EndTime int64 `json:"endTime"`

	// Exchangeable 是否可兑换
	// 1: 可兑换, 0: 不可兑换
	Exchangeable int64 `json:"exchangeable"`

	// MaxBuyPrice 最大买入价格限制
	MaxBuyPrice float64 `json:"maxBuyPrice"`

	// MaxVolume 最大成交量限制
	MaxVolume float64 `json:"maxVolume"`

	// MinVolume 最小成交量限制
	MinVolume float64 `json:"minVolume"`

	// PublishAmount 发行总量
	PublishAmount float64 `json:"publishAmount"`

	// PublishPrice 发行价格
	PublishPrice float64 `json:"publishPrice"`

	// PublishType 发行类型
	PublishType int64 `json:"publishType"`

	// RobotType 做市机器人类型
	RobotType int64 `json:"robotType"`

	// StartTime 交易开始时间（时间戳）
	StartTime int64 `json:"startTime"`

	// Visible 是否对用户可见
	// 1: 可见, 0: 隐藏
	Visible int64 `json:"visible"`

	// Zone 交易区域标识
	Zone int64 `json:"zone"`

	// CurrentTime 当前服务器时间（时间戳）
	// 用于前端同步时间显示
	CurrentTime int64 `json:"currentTime"`

	// MarketEngineStatus 市场引擎状态
	// 表示撮合引擎的运行状态
	MarketEngineStatus int `json:"marketEngineStatus"`

	// EngineStatus 交易引擎状态
	EngineStatus int `json:"engineStatus"`

	// ExEngineStatus 扩展引擎状态
	ExEngineStatus int `json:"exEngineStatus"`
}

// Coin 是 market API 返回的公开币种元数据模型。
// 包含币种的完整信息，用于币种详情展示、充值提现等功能。
//
// 数据来源：
//   - 从数据库 coin 表读取
//   - 包含币种配置、充提规则、钱包地址等信息
//
// 使用场景：
//   - 币种列表展示
//   - 充值/提现页面
//   - 资产管理
type Coin struct {
	// ID 币种唯一标识
	ID int `json:"id"`

	// Name 币种全称，如 "Bitcoin"
	Name string `json:"name"`

	// CanAutoWithdraw 是否允许自动提现
	// 1: 允许, 0: 禁止
	CanAutoWithdraw int `json:"canAutoWithdraw"`

	// CanRecharge 是否允许充值
	// 1: 允许, 0: 禁止
	CanRecharge int `json:"canRecharge"`

	// CanTransfer 是否允许转账
	// 1: 允许, 0: 禁止
	CanTransfer int `json:"canTransfer"`

	// CanWithdraw 是否允许提现
	// 1: 允许, 0: 禁止
	CanWithdraw int `json:"canWithdraw"`

	// CNYRate 对人民币的汇率
	// 用于显示人民币计价的资产价值
	CNYRate float64 `json:"cnyRate"`

	// EnableRPC 是否启用 RPC 接口
	// 用于与区块链节点交互
	EnableRPC int `json:"enableRpc"`

	// IsPlatformCoin 是否为平台币
	// 1: 是平台币, 0: 非平台币
	IsPlatformCoin int `json:"isPlatformCoin"`

	// MaxTxFee 最大交易手续费
	MaxTxFee float64 `json:"maxTxFee"`

	// MaxWithdrawAmount 单次最大提现金额
	MaxWithdrawAmount float64 `json:"maxWithdrawAmount"`

	// MinTxFee 最小交易手续费
	MinTxFee float64 `json:"minTxFee"`

	// MinWithdrawAmount 单次最小提现金额
	MinWithdrawAmount float64 `json:"minWithdrawAmount"`

	// NameCN 币种中文名称，如 "比特币"
	NameCN string `json:"nameCn"`

	// Sort 排序权重，数值越小越靠前
	Sort int `json:"sort"`

	// Status 币种状态
	// 通常: 0-禁用, 1-启用
	Status int `json:"status"`

	// Unit 币种单位代码，如 "BTC", "ETH", "USDT"
	Unit string `json:"unit"`

	// USDTRate 对 USDT 的汇率
	USDTRate float64 `json:"usdRate"`

	// WithdrawThreshold 提现阈值
	// 超过此金额需要人工审核
	WithdrawThreshold float64 `json:"withdrawThreshold"`

	// HasLegal 是否有法币通道
	// 1: 有, 0: 无
	HasLegal int `json:"hasLegal"`

	// ColdWalletAddress 冷钱包地址
	// 用于大额资金存储
	ColdWalletAddress string `json:"coldWalletAddress"`

	// MinerFee 矿工费/网络手续费
	MinerFee float64 `json:"minerFee"`

	// WithdrawScale 提现精度（小数位数）
	WithdrawScale int `json:"withdrawScale"`

	// AccountType 账户类型
	// 区分不同区块链网络的账户类型
	AccountType int `json:"accountType"`

	// DepositAddress 默认充值地址
	// 用于展示用户的充值地址
	DepositAddress string `json:"depositAddress"`

	// InfoLink 币种信息链接
	// 指向币种详细介绍页面
	InfoLink string `json:"infolink"`

	// Information 币种简介
	// 包含币种的基本介绍信息
	Information string `json:"information"`

	// MinRechargeAmount 最小充值金额
	// 低于此金额的充值可能不会入账
	MinRechargeAmount float64 `json:"minRechargeAmount"`
}

// HistoryKline 是 K 线历史数据的响应结构。
// 该结构保持与旧 API 的兼容性，响应格式为原始的 OHLCV 数据列表。
//
// 设计说明：
//   - 使用 [][]any 而非结构体数组，是为了保持与旧 API 的格式一致
//   - 每个内部数组包含 6 个元素: [time, open, high, low, close, volume]
//   - 前端可以直接使用此格式渲染 K 线图表
//
// 响应示例:
//
//	{
//	  "list": [
//	    [1704067200000, 42000.0, 42500.0, 41800.0, 42300.0, 123.45],
//	    [1704070800000, 42300.0, 42800.0, 42200.0, 42700.0, 234.56]
//	  ]
//	}
type HistoryKline struct {
	// List K 线数据列表
	// 每个元素为一个 K 线数据点: [时间戳, 开盘价, 最高价, 最低价, 收盘价, 成交量]
	//
	// 数据格式:
	//   - 时间戳: 毫秒级 Unix 时间戳
	//   - 价格: 以计价货币计
	//   - 成交量: 以基础货币计
	List [][]any `json:"list"`
}