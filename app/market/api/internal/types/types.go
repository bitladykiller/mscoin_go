package types

// RateRequest is the request payload for fiat-rate queries.
type RateRequest struct {
	Unit string `path:"unit" json:"unit"`
	IP   string `json:"ip,optional"`
}

// RateResponse keeps the original simple response shape for rate lookup.
type RateResponse struct {
	Rate float64 `json:"rate"`
}

// MarketReq is reused by the HTTP endpoints that proxy market queries to the
// RPC layer. The field names intentionally stay aligned with the old API so the
// frontend contract remains stable.
type MarketReq struct {
	IP         string `json:"ip,optional"`
	Symbol     string `json:"symbol,optional" form:"symbol,optional" path:"symbol,optional"`
	Unit       string `json:"unit,optional" form:"unit,optional"`
	From       int64  `json:"from,optional" form:"from,optional"`
	To         int64  `json:"to,optional" form:"to,optional"`
	Resolution string `json:"resolution,optional" form:"resolution,optional"`
}

// CoinThumbResp is the public market snapshot model returned to web clients.
type CoinThumbResp struct {
	Symbol       string    `json:"symbol"`
	Open         float64   `json:"open"`
	High         float64   `json:"high"`
	Low          float64   `json:"low"`
	Close        float64   `json:"close"`
	Chg          float64   `json:"chg"`
	Change       float64   `json:"change"`
	Volume       float64   `json:"volume"`
	Turnover     float64   `json:"turnover"`
	LastDayClose float64   `json:"lastDayClose"`
	USDTRate     float64   `json:"usdRate"`
	BaseUSDTRate float64   `json:"baseUsdRate"`
	Zone         int       `json:"zone"`
	Trend        []float64 `json:"trend,optional"`
}

// ExchangeCoinResp is the public model used for trading-pair metadata.
type ExchangeCoinResp struct {
	ID                 int64   `json:"id"`
	Symbol             string  `json:"symbol"`
	BaseCoinScale      int64   `json:"baseCoinScale"`
	BaseSymbol         string  `json:"baseSymbol"`
	CoinScale          int64   `json:"coinScale"`
	CoinSymbol         string  `json:"coinSymbol"`
	Enable             int64   `json:"enable"`
	Fee                float64 `json:"fee"`
	Sort               int64   `json:"sort"`
	EnableMarketBuy    int64   `json:"enableMarketBuy"`
	EnableMarketSell   int64   `json:"enableMarketSell"`
	MinSellPrice       float64 `json:"minSellPrice"`
	Flag               int64   `json:"flag"`
	MaxTradingOrder    int64   `json:"maxTradingOrder"`
	MaxTradingTime     int64   `json:"maxTradingTime"`
	MinTurnover        float64 `json:"minTurnover"`
	ClearTime          int64   `json:"clearTime"`
	EndTime            int64   `json:"endTime"`
	Exchangeable       int64   `json:"exchangeable"`
	MaxBuyPrice        float64 `json:"maxBuyPrice"`
	MaxVolume          float64 `json:"maxVolume"`
	MinVolume          float64 `json:"minVolume"`
	PublishAmount      float64 `json:"publishAmount"`
	PublishPrice       float64 `json:"publishPrice"`
	PublishType        int64   `json:"publishType"`
	RobotType          int64   `json:"robotType"`
	StartTime          int64   `json:"startTime"`
	Visible            int64   `json:"visible"`
	Zone               int64   `json:"zone"`
	CurrentTime        int64   `json:"currentTime"`
	MarketEngineStatus int     `json:"marketEngineStatus"`
	EngineStatus       int     `json:"engineStatus"`
	ExEngineStatus     int     `json:"exEngineStatus"`
}

// Coin is the public coin metadata model returned by the market API.
type Coin struct {
	ID                int     `json:"id"`
	Name              string  `json:"name"`
	CanAutoWithdraw   int     `json:"canAutoWithdraw"`
	CanRecharge       int     `json:"canRecharge"`
	CanTransfer       int     `json:"canTransfer"`
	CanWithdraw       int     `json:"canWithdraw"`
	CNYRate           float64 `json:"cnyRate"`
	EnableRPC         int     `json:"enableRpc"`
	IsPlatformCoin    int     `json:"isPlatformCoin"`
	MaxTxFee          float64 `json:"maxTxFee"`
	MaxWithdrawAmount float64 `json:"maxWithdrawAmount"`
	MinTxFee          float64 `json:"minTxFee"`
	MinWithdrawAmount float64 `json:"minWithdrawAmount"`
	NameCN            string  `json:"nameCn"`
	Sort              int     `json:"sort"`
	Status            int     `json:"status"`
	Unit              string  `json:"unit"`
	USDTRate          float64 `json:"usdRate"`
	WithdrawThreshold float64 `json:"withdrawThreshold"`
	HasLegal          int     `json:"hasLegal"`
	ColdWalletAddress string  `json:"coldWalletAddress"`
	MinerFee          float64 `json:"minerFee"`
	WithdrawScale     int     `json:"withdrawScale"`
	AccountType       int     `json:"accountType"`
	DepositAddress    string  `json:"depositAddress"`
	InfoLink          string  `json:"infolink"`
	Information       string  `json:"information"`
	MinRechargeAmount float64 `json:"minRechargeAmount"`
}

// HistoryKline preserves the old API handler behavior where the response wraps
// a raw list of OHLCV rows instead of an object with field names on every row.
type HistoryKline struct {
	List [][]any `json:"list"`
}
