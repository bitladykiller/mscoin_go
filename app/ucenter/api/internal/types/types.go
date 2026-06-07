package types

type CaptchaReq struct {
	Server string `json:"server"`
	Token  string `json:"token"`
}

type Request struct {
	Username     string      `json:"username,optional"`
	Password     string      `json:"password,optional"`
	Captcha      *CaptchaReq `json:"captcha,optional"`
	Phone        string      `json:"phone,optional"`
	Promotion    string      `json:"promotion,optional"`
	Code         string      `json:"code,optional"`
	Country      string      `json:"country,optional"`
	SuperPartner string      `json:"superPartner,optional"`
	IP           string      `json:"ip,optional"`
}

type Response struct {
	Message string `json:"message"`
}

type CodeRequest struct {
	Phone   string `json:"phone,optional"`
	Country string `json:"country,optional"`
}

type CodeResponse struct{}

type LoginReq struct {
	Username string      `json:"username"`
	Password string      `json:"password"`
	Captcha  *CaptchaReq `json:"captcha,optional"`
	IP       string      `json:"ip,optional"`
}

type LoginRes struct {
	Username      string `json:"username"`
	Token         string `json:"token"`
	MemberLevel   string `json:"memberLevel"`
	RealName      string `json:"realName"`
	Country       string `json:"country"`
	Avatar        string `json:"avatar"`
	PromotionCode string `json:"promotionCode"`
	Id            int64  `json:"id"`
	LoginCount    int    `json:"loginCount"`
	SuperPartner  string `json:"superPartner"`
	MemberRate    int    `json:"memberRate"`
}

type AssetReq struct {
	CoinName  string `json:"coinName,optional" path:"coinName,optional"`
	IP        string `json:"ip,optional"`
	Unit      string `json:"unit,optional" form:"unit,optional"`
	PageNo    int    `json:"pageNo,optional" form:"pageNo,optional"`
	PageSize  int    `json:"pageSize,optional" form:"pageSize,optional"`
	StartTime string `json:"startTime,optional" form:"startTime,optional"`
	EndTime   string `json:"endTime,optional" form:"endTime,optional"`
	Symbol    string `json:"symbol,optional" form:"symbol,optional"`
	Type      string `json:"type,optional" form:"type,optional"`
}

type Coin struct {
	Id                int32   `json:"id"`
	Name              string  `json:"name"`
	CanAutoWithdraw   int32   `json:"canAutoWithdraw"`
	CanRecharge       int32   `json:"canRecharge"`
	CanTransfer       int32   `json:"canTransfer"`
	CanWithdraw       int32   `json:"canWithdraw"`
	CnyRate           float64 `json:"cnyRate"`
	EnableRpc         int32   `json:"enableRpc"`
	IsPlatformCoin    int32   `json:"isPlatformCoin"`
	MaxTxFee          float64 `json:"maxTxFee"`
	MaxWithdrawAmount float64 `json:"maxWithdrawAmount"`
	MinTxFee          float64 `json:"minTxFee"`
	MinWithdrawAmount float64 `json:"minWithdrawAmount"`
	NameCn            string  `json:"nameCn"`
	Sort              int32   `json:"sort"`
	Status            int32   `json:"status"`
	Unit              string  `json:"unit"`
	UsdRate           float64 `json:"usdRate"`
	WithdrawThreshold float64 `json:"withdrawThreshold"`
	HasLegal          int32   `json:"hasLegal"`
	ColdWalletAddress string  `json:"coldWalletAddress"`
	MinerFee          float64 `json:"minerFee"`
	WithdrawScale     int32   `json:"withdrawScale"`
	AccountType       int32   `json:"accountType"`
	DepositAddress    string  `json:"depositAddress"`
	Infolink          string  `json:"infolink"`
	Information       string  `json:"information"`
	MinRechargeAmount float64 `json:"minRechargeAmount"`
}

type MemberWallet struct {
	Id             int64   `json:"id"`
	Address        string  `json:"address"`
	Balance        float64 `json:"balance"`
	FrozenBalance  float64 `json:"frozenBalance"`
	ReleaseBalance float64 `json:"releaseBalance"`
	IsLock         int32   `json:"isLock"`
	MemberId       int64   `json:"memberId"`
	Version        int32   `json:"version"`
	Coin           Coin    `json:"coin"`
	ToReleased     float64 `json:"toReleased"`
}

type MemberTransaction struct {
	Id          int64   `json:"id"`
	Address     string  `json:"address"`
	Amount      float64 `json:"amount"`
	CreateTime  string  `json:"createTime"`
	Fee         float64 `json:"fee"`
	Flag        int32   `json:"flag"`
	MemberId    int64   `json:"memberId"`
	Symbol      string  `json:"symbol"`
	Type        string  `json:"type"`
	DiscountFee string  `json:"discountFee"`
	RealFee     string  `json:"realFee"`
}

type ApproveReq struct{}

type MemberSecurity struct {
	Username             string `json:"username"`
	Id                   int64  `json:"id"`
	CreateTime           string `json:"createTime"`
	RealVerified         string `json:"realVerified"`
	EmailVerified        string `json:"emailVerified"`
	PhoneVerified        string `json:"phoneVerified"`
	LoginVerified        string `json:"loginVerified"`
	FundsVerified        string `json:"fundsVerified"`
	RealAuditing         string `json:"realAuditing"`
	MobilePhone          string `json:"mobilePhone"`
	Email                string `json:"email"`
	RealName             string `json:"realName"`
	RealNameRejectReason string `json:"realNameRejectReason"`
	IdCard               string `json:"idCard"`
	Avatar               string `json:"avatar"`
	AccountVerified      string `json:"accountVerified"`
}

// WithdrawReq keeps the legacy withdraw form payload intact so the refactored
// API can continue serving the existing frontend contract without requiring any
// request-shape changes on the caller side.
type WithdrawReq struct {
	Unit       string  `json:"unit,optional" form:"unit,optional"`
	Address    string  `json:"address,optional" form:"address,optional"`
	Amount     float64 `json:"amount,optional" form:"amount,optional"`
	Fee        float64 `json:"fee,optional" form:"fee,optional"`
	JyPassword string  `json:"jyPassword,optional" form:"jyPassword,optional"`
	Code       string  `json:"code,optional" form:"code,optional"`
	Page       int     `json:"page,optional" form:"page,optional"`
	PageSize   int     `json:"pageSize,optional" form:"pageSize,optional"`
}

// AddressSimple mirrors the legacy address-book projection returned to the
// withdraw page.
type AddressSimple struct {
	Remark  string `json:"remark"`
	Address string `json:"address"`
}

// WithdrawWalletInfo is the aggregated view required by the legacy withdraw UI.
// It combines market metadata, member wallet balance, and saved address book
// entries into one response object.
type WithdrawWalletInfo struct {
	Unit            string          `json:"unit"`
	Threshold       float64         `json:"threshold"`
	MinAmount       float64         `json:"minAmount"`
	MaxAmount       float64         `json:"maxAmount"`
	MinTxFee        float64         `json:"minTxFee"`
	MaxTxFee        float64         `json:"maxTxFee"`
	NameCn          string          `json:"nameCn"`
	Name            string          `json:"name"`
	Balance         float64         `json:"balance"`
	CanAutoWithdraw string          `json:"canAutoWithdraw"`
	WithdrawScale   int32           `json:"withdrawScale"`
	AccountType     int32           `json:"accountType"`
	Addresses       []AddressSimple `json:"addresses"`
}

// WithdrawRecord mirrors the historical withdraw detail envelope returned to
// clients so pagination screens remain backwards compatible.
type WithdrawRecord struct {
	Id                int64   `json:"id"`
	MemberId          int64   `json:"memberId"`
	Coin              Coin    `json:"coin"`
	TotalAmount       float64 `json:"totalAmount"`
	Fee               float64 `json:"fee"`
	ArrivedAmount     float64 `json:"arrivedAmount"`
	Address           string  `json:"address"`
	Remark            string  `json:"remark"`
	TransactionNumber string  `json:"transactionNumber"`
	CanAutoWithdraw   int32   `json:"canAutoWithdraw"`
	IsAuto            int32   `json:"isAuto"`
	Status            int32   `json:"status"`
	CreateTime        string  `json:"createTime"`
	DealTime          string  `json:"dealTime"`
}
