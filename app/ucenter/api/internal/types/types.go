// Package types 定义 ucenter-api 服务的请求和响应数据结构。
//
// 该包包含所有 HTTP 接口的输入输出类型定义，是 API 层与 Logic 层之间的数据契约。
// 结构体标签使用 go-zero 的验证标签和 JSON 序列化标签。
//
// 数据结构分类：
//   - 认证相关：CaptchaReq、LoginReq、LoginRes 等
//   - 钱包资产：AssetReq、MemberWallet、Coin 等
//   - 交易记录：MemberTransaction
//   - 提现相关：WithdrawReq、WithdrawWalletInfo、WithdrawRecord 等
//   - 安全设置：MemberSecurity
package types

// CaptchaReq 是验证码验证请求数据。
//
// 用于注册和登录时的验证码校验，支持第三方验证码服务（如 Google reCAPTCHA）。
type CaptchaReq struct {
	// Server 是验证码服务服务器地址。
	Server string `json:"server"`

	// Token 是验证码响应令牌，由前端从验证码服务获取。
	Token string `json:"token"`
}

// Request 是通用的用户请求结构，主要用于注册接口。
//
// 该结构支持多种注册方式：用户名密码注册、手机号注册、邀请码注册等。
type Request struct {
	// Username 是用户名，用于用户名密码注册方式。
	Username     string      `json:"username,optional"`

	// Password 是用户密码（明文），服务端会进行加密存储。
	Password     string      `json:"password,optional"`

	// Captcha 是验证码验证数据，用于防止机器人注册/登录。
	Captcha      *CaptchaReq `json:"captcha,optional"`

	// Phone 是手机号码，用于手机号注册方式。
	Phone        string      `json:"phone,optional"`

	// Promotion 是邀请码，用于记录邀请关系。
	Promotion    string      `json:"promotion,optional"`

	// Code 是短信验证码，用于手机号注册验证。
	Code         string      `json:"code,optional"`

	// Country 是国家代码，如 "CN"、"US"，用于国际手机号支持。
	Country      string      `json:"country,optional"`

	// SuperPartner 是超级合伙人标识，用于合伙人体系。
	SuperPartner string      `json:"superPartner,optional"`

	// IP 是客户端 IP 地址，用于安全审计和风控。
	IP           string      `json:"ip,optional"`
}

// Response 是通用的响应结构。
type Response struct {
	// Message 是响应消息，通常用于错误提示。
	Message string `json:"message"`
}

// CodeRequest 是发送短信验证码的请求结构。
type CodeRequest struct {
	// Phone 是接收验证码的手机号码。
	Phone   string `json:"phone,optional"`

	// Country 是手机号所属国家代码。
	Country string `json:"country,optional"`
}

// CodeResponse 是发送验证码的响应结构。
// 发送成功时返回空对象，客户端通过 HTTP 状态码判断结果。
type CodeResponse struct{}

// LoginReq 是用户登录请求结构。
//
// 登录流程：
//  1. 前端提交用户名、密码、验证码
//  2. API 层调用 ucenter-rpc 的 Login 服务验证
//  3. 验证通过后返回 JWT Token 和用户信息
type LoginReq struct {
	// Username 是用户名。
	Username string      `json:"username"`

	// Password 是用户密码。
	Password string      `json:"password"`

	// Captcha 是验证码验证数据。
	Captcha  *CaptchaReq `json:"captcha,optional"`

	// IP 是客户端 IP 地址，用于安全审计和异地登录检测。
	IP       string      `json:"ip,optional"`
}

// LoginRes 是用户登录响应结构。
//
// 包含用户基本信息和 JWT Token，前端应保存 Token 用于后续请求认证。
type LoginRes struct {
	// Username 是用户名。
	Username      string `json:"username"`

	// Token 是 JWT Token，前端需在后续请求的 x-auth-token 头部携带。
	Token         string `json:"token"`

	// MemberLevel 是会员等级标识。
	MemberLevel   string `json:"memberLevel"`

	// RealName 是用户真实姓名（实名认证后）。
	RealName      string `json:"realName"`

	// Country 是用户所在国家。
	Country       string `json:"country"`

	// Avatar 是用户头像 URL。
	Avatar        string `json:"avatar"`

	// PromotionCode 是用户的邀请码，可用于邀请新用户。
	PromotionCode string `json:"promotionCode"`

	// Id 是用户 ID。
	Id            int64  `json:"id"`

	// LoginCount 是用户登录次数。
	LoginCount    int    `json:"loginCount"`

	// SuperPartner 是超级合伙人标识。
	SuperPartner  string `json:"superPartner"`

	// MemberRate 是会员费率等级。
	MemberRate    int    `json:"memberRate"`
}

// AssetReq 是资产查询请求结构。
//
// 用于查询钱包余额、交易记录等资产相关信息，支持分页和条件筛选。
type AssetReq struct {
	// CoinName 是币种名称（如 BTC、ETH），用于查询单个币种钱包。
	// 通过 URL 路径传递：/uc/asset/wallet/:coinName
	CoinName  string `json:"coinName,optional" path:"coinName,optional"`

	// IP 是客户端 IP 地址，用于安全审计。
	IP        string `json:"ip,optional"`

	// Unit 是币种单位，如 "USDT"、"BTC"。
	Unit      string `json:"unit,optional" form:"unit,optional"`

	// PageNo 是页码，从 1 开始。
	PageNo    int    `json:"pageNo,optional" form:"pageNo,optional"`

	// PageSize 是每页数量，默认 10。
	PageSize  int    `json:"pageSize,optional" form:"pageSize,optional"`

	// StartTime 是查询起始时间（格式：yyyy-MM-dd HH:mm:ss）。
	StartTime string `json:"startTime,optional" form:"startTime,optional"`

	// EndTime 是查询结束时间（格式：yyyy-MM-dd HH:mm:ss）。
	EndTime   string `json:"endTime,optional" form:"endTime,optional"`

	// Symbol 是交易对符号，用于交易记录筛选。
	Symbol    string `json:"symbol,optional" form:"symbol,optional"`

	// Type 是交易类型，用于交易记录筛选（如充值、提现、转账等）。
	Type      string `json:"type,optional" form:"type,optional"`
}

// Coin 是币种信息结构。
//
// 包含币种的完整配置信息，如提现限额、手续费、充值提现开关等。
// 该数据来自 market-rpc 服务。
type Coin struct {
	// Id 是币种 ID。
	Id                int32   `json:"id"`

	// Name 是币种名称（如 Bitcoin）。
	Name              string  `json:"name"`

	// CanAutoWithdraw 是否支持自动提现（0: 支持，1: 不支持）。
	CanAutoWithdraw   int32   `json:"canAutoWithdraw"`

	// CanRecharge 是否开放充值（0: 开放，1: 关闭）。
	CanRecharge       int32   `json:"canRecharge"`

	// CanTransfer 是否支持转账（0: 支持，1: 不支持）。
	CanTransfer       int32   `json:"canTransfer"`

	// CanWithdraw 是否开放提现（0: 开放，1: 关闭）。
	CanWithdraw       int32   `json:"canWithdraw"`

	// CnyRate 是人民币汇率。
	CnyRate           float64 `json:"cnyRate"`

	// EnableRpc 是否启用 RPC 同步（0: 启用，1: 禁用）。
	EnableRpc         int32   `json:"enableRpc"`

	// IsPlatformCoin 是否为平台币（0: 是，1: 否）。
	IsPlatformCoin    int32   `json:"isPlatformCoin"`

	// MaxTxFee 是最大矿工费。
	MaxTxFee          float64 `json:"maxTxFee"`

	// MaxWithdrawAmount 是单次最大提现金额。
	MaxWithdrawAmount float64 `json:"maxWithdrawAmount"`

	// MinTxFee 是最小矿工费。
	MinTxFee          float64 `json:"minTxFee"`

	// MinWithdrawAmount 是单次最小提现金额。
	MinWithdrawAmount float64 `json:"minWithdrawAmount"`

	// NameCn 是币种中文名称。
	NameCn            string  `json:"nameCn"`

	// Sort 是排序权重。
	Sort              int32   `json:"sort"`

	// Status 是币种状态（0: 正常，1: 停用）。
	Status            int32   `json:"status"`

	// Unit 是币种单位（如 BTC、ETH）。
	Unit              string  `json:"unit"`

	// UsdRate 是美元汇率。
	UsdRate           float64 `json:"usdRate"`

	// WithdrawThreshold 是提现阈值（余额需大于此值才能提现）。
	WithdrawThreshold float64 `json:"withdrawThreshold"`

	// HasLegal 是否有法币通道（0: 有，1: 无）。
	HasLegal          int32   `json:"hasLegal"`

	// ColdWalletAddress 是冷钱包地址。
	ColdWalletAddress string  `json:"coldWalletAddress"`

	// MinerFee 是矿工费。
	MinerFee          float64 `json:"minerFee"`

	// WithdrawScale 是提现精度（小数位数）。
	WithdrawScale     int32   `json:"withdrawScale"`

	// AccountType 是账户类型。
	AccountType       int32   `json:"accountType"`

	// DepositAddress 是默认充值地址（用于展示）。
	DepositAddress    string  `json:"depositAddress"`

	// Infolink 是币种信息链接。
	Infolink          string  `json:"infolink"`

	// Information 是币种简介。
	Information       string  `json:"information"`

	// MinRechargeAmount 是最小充值金额。
	MinRechargeAmount float64 `json:"minRechargeAmount"`
}

// MemberWallet 是会员钱包结构。
//
// 表示用户持有的单个币种钱包信息，包含余额、冻结余额、充值地址等。
type MemberWallet struct {
	// Id 是钱包 ID。
	Id             int64   `json:"id"`

	// Address 是充值地址，用户可向此地址充币。
	Address        string  `json:"address"`

	// Balance 是可用余额。
	Balance        float64 `json:"balance"`

	// FrozenBalance 是冻结余额（如提现中、锁定等）。
	FrozenBalance  float64 `json:"frozenBalance"`

	// ReleaseBalance 是已释放余额（用于锁仓释放）。
	ReleaseBalance float64 `json:"releaseBalance"`

	// IsLock 是钱包锁定状态（0: 正常，1: 锁定）。
	IsLock         int32   `json:"isLock"`

	// MemberId 是所属会员 ID。
	MemberId       int64   `json:"memberId"`

	// Version 是版本号（用于乐观锁）。
	Version        int32   `json:"version"`

	// Coin 是关联的币种信息。
	Coin           Coin    `json:"coin"`

	// ToReleased 是待释放余额（锁仓奖励等）。
	ToReleased     float64 `json:"toReleased"`
}

// MemberTransaction 是会员交易记录结构。
//
// 记录用户的充值、提现、转账等交易历史。
type MemberTransaction struct {
	// Id 是交易 ID。
	Id          int64   `json:"id"`

	// Address 是交易对方地址（充值来源地址或提现目标地址）。
	Address     string  `json:"address"`

	// Amount 是交易金额。
	Amount      float64 `json:"amount"`

	// CreateTime 是交易时间。
	CreateTime  string  `json:"createTime"`

	// Fee 是手续费。
	Fee         float64 `json:"fee"`

	// Flag 是交易标志（充值/提现方向等）。
	Flag        int32   `json:"flag"`

	// MemberId 是所属会员 ID。
	MemberId    int64   `json:"memberId"`

	// Symbol 是交易币种符号。
	Symbol      string  `json:"symbol"`

	// Type 是交易类型（充值、提现、转账等）。
	Type        string  `json:"type"`

	// DiscountFee 是优惠手续费。
	DiscountFee string  `json:"discountFee"`

	// RealFee 是实际手续费。
	RealFee     string  `json:"realFee"`
}

// ApproveReq 是审批请求的空结构。
// 用于安全设置等无需参数的接口。
type ApproveReq struct{}

// MemberSecurity 是会员安全设置信息。
//
// 展示用户的安全认证状态，包括实名认证、手机认证、邮箱认证、资金密码等。
type MemberSecurity struct {
	// Username 是用户名。
	Username             string `json:"username"`

	// Id 是会员 ID。
	Id                   int64  `json:"id"`

	// CreateTime 是注册时间。
	CreateTime           string `json:"createTime"`

	// RealVerified 是实名认证状态（"true"/"false"）。
	RealVerified         string `json:"realVerified"`

	// EmailVerified 是邮箱认证状态（"true"/"false"）。
	EmailVerified        string `json:"emailVerified"`

	// PhoneVerified 是手机认证状态（"true"/"false"）。
	PhoneVerified        string `json:"phoneVerified"`

	// LoginVerified 是登录验证状态（"true"/"false"）。
	LoginVerified        string `json:"loginVerified"`

	// FundsVerified 是资金密码设置状态（"true"/"false"）。
	FundsVerified        string `json:"fundsVerified"`

	// RealAuditing 是实名审核中状态（"true"/"false"）。
	RealAuditing         string `json:"realAuditing"`

	// MobilePhone 是已认证手机号（脱敏显示）。
	MobilePhone          string `json:"mobilePhone"`

	// Email 是已认证邮箱。
	Email                string `json:"email"`

	// RealName 是已认证真实姓名。
	RealName             string `json:"realName"`

	// RealNameRejectReason 是实名认证拒绝原因。
	RealNameRejectReason string `json:"realNameRejectReason"`

	// IdCard 是身份证号（脱敏显示，仅显示前两位）。
	IdCard               string `json:"idCard"`

	// Avatar 是用户头像 URL。
	Avatar               string `json:"avatar"`

	// AccountVerified 是账户验证状态（绑定银行卡/支付宝/微信）。
	AccountVerified      string `json:"accountVerified"`
}

// WithdrawReq 是提现相关请求结构。
//
// 该结构用于多个提现相关的接口：
//   - 提现申请：需要填写 unit、address、amount、fee、jyPassword、code
//   - 提现记录查询：使用 page、pageSize 进行分页
//
// 保持遗留提现表单负载的完整性，以便重构后的 API
// 能够继续服务现有的前端契约，而无需调用方做任何请求格式的变更。
type WithdrawReq struct {
	// Unit 是币种单位（如 USDT、BTC）。
	Unit       string  `json:"unit,optional" form:"unit,optional"`

	// Address 是提现目标地址。
	Address    string  `json:"address,optional" form:"address,optional"`

	// Amount 是提现金额。
	Amount     float64 `json:"amount,optional" form:"amount,optional"`

	// Fee 是矿工费（手续费）。
	Fee        float64 `json:"fee,optional" form:"fee,optional"`

	// JyPassword 是资金密码，用于验证提现操作权限。
	JyPassword string  `json:"jyPassword,optional" form:"jyPassword,optional"`

	// Code 是短信验证码，用于二次验证。
	Code       string  `json:"code,optional" form:"code,optional"`

	// Page 是页码（用于提现记录查询）。
	Page       int     `json:"page,optional" form:"page,optional"`

	// PageSize 是每页数量（用于提现记录查询）。
	PageSize   int     `json:"pageSize,optional" form:"pageSize,optional"`
}

// AddressSimple 是提现地址簿的简化结构。
//
// 映射提现页面返回的遗留地址簿投影数据，用于展示用户保存的常用提现地址。
type AddressSimple struct {
	// Remark 是地址备注名称。
	Remark  string `json:"remark"`

	// Address 是提现地址。
	Address string `json:"address"`
}

// WithdrawWalletInfo 是提现钱包信息聚合结构。
//
// 该结构是遗留提现 UI 所需的聚合视图，将以下信息合并为一个响应对象：
//   - 币种基本信息（名称、限额、手续费等）：来自 market-rpc
//   - 会员钱包余额：来自 ucenter-rpc AssetClient
//   - 已保存的提现地址列表：来自 ucenter-rpc WithdrawClient
type WithdrawWalletInfo struct {
	// Unit 是币种单位。
	Unit            string          `json:"unit"`

	// Threshold 是提现阈值（余额需大于此值才能提现）。
	Threshold       float64         `json:"threshold"`

	// MinAmount 是单次最小提现金额。
	MinAmount       float64         `json:"minAmount"`

	// MaxAmount 是单次最大提现金额。
	MaxAmount       float64         `json:"maxAmount"`

	// MinTxFee 是最小矿工费。
	MinTxFee        float64         `json:"minTxFee"`

	// MaxTxFee 是最大矿工费。
	MaxTxFee        float64         `json:"maxTxFee"`

	// NameCn 是币种中文名称。
	NameCn          string          `json:"nameCn"`

	// Name 是币种英文名称。
	Name            string          `json:"name"`

	// Balance 是用户该币种的可用余额。
	Balance         float64         `json:"balance"`

	// CanAutoWithdraw 是否支持自动提现（"true"/"false"）。
	CanAutoWithdraw string          `json:"canAutoWithdraw"`

	// WithdrawScale 是提现精度（小数位数）。
	WithdrawScale   int32           `json:"withdrawScale"`

	// AccountType 是账户类型。
	AccountType     int32           `json:"accountType"`

	// Addresses 是用户已保存的提现地址列表。
	Addresses       []AddressSimple `json:"addresses"`
}

// WithdrawRecord 是提现记录结构。
//
// 映射返回给客户端的历史提现详情数据结构，
// 以保持分页界面的向后兼容性。
type WithdrawRecord struct {
	// Id 是提现记录 ID。
	Id                int64   `json:"id"`

	// MemberId 是会员 ID。
	MemberId          int64   `json:"memberId"`

	// Coin 是提现币种信息。
	Coin              Coin    `json:"coin"`

	// TotalAmount 是提现总金额。
	TotalAmount       float64 `json:"totalAmount"`

	// Fee 是手续费。
	Fee               float64 `json:"fee"`

	// ArrivedAmount 是到账金额（总金额 - 手续费）。
	ArrivedAmount     float64 `json:"arrivedAmount"`

	// Address 是提现目标地址。
	Address           string  `json:"address"`

	// Remark 是提现备注。
	Remark            string  `json:"remark"`

	// TransactionNumber 是区块链交易哈希。
	TransactionNumber string  `json:"transactionNumber"`

	// CanAutoWithdraw 是否支持自动提现（0: 支持，1: 不支持）。
	CanAutoWithdraw   int32   `json:"canAutoWithdraw"`

	// IsAuto 是否为自动提现（0: 是，1: 否）。
	IsAuto            int32   `json:"isAuto"`

	// Status 是提现状态（0: 待审核，1: 已审核，2: 已完成，3: 已拒绝等）。
	Status            int32   `json:"status"`

	// CreateTime 是创建时间。
	CreateTime        string  `json:"createTime"`

	// DealTime 是处理时间。
	DealTime          string  `json:"dealTime"`
}
