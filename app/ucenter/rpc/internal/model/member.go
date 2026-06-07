// Package model 定义 ucenter 服务的领域模型。
//
// 本包包含会员、钱包、交易、提现记录等核心业务实体。
// 模型采用贫血模型设计，主要职责是数据承载和简单的业务规则封装。
// 复杂的业务逻辑由 domain/service 层处理。
//
// 模型与数据库表的映射关系：
//   - Member -> member 表
//   - MemberWallet -> member_wallet 表
//   - MemberTransaction -> member_transaction 表
//   - MemberAddress -> member_address 表
//   - WithdrawRecord -> withdraw_record 表
package model

import "time"

// Member 会员模型
// 存储会员的基本信息、认证状态、钱包关联等数据
//
// 会员是 MSCoin 平台的核心实体，每个会员拥有：
//   - 唯一的手机号作为登录账号
//   - 多个币种钱包（通过 MemberWallet 关联）
//   - 提现地址簿（通过 MemberAddress 关联）
//   - 交易记录（通过 MemberTransaction 关联）
//
// 会员状态与等级：
//   - MemberLevel：普通会员 -> 实名会员 -> 认证商家
//   - SuperPartner：普通会员 -> 超级合伙人 -> P 级超级合伙人
//   - Status：正常状态 -> 违规状态
type Member struct {
	Id                         int64   `db:"id" gorm:"column:id"`                                   // 会员 ID，自增主键
	AliNo                      string  `db:"ali_no" gorm:"column:ali_no"`                           // 支付宝账号，用于收款
	QrCodeUrl                  string  `db:"qr_code_url" gorm:"column:qr_code_url"`                 // 二维码 URL，收款码图片
	AppealSuccessTimes         int64   `db:"appeal_success_times" gorm:"column:appeal_success_times"` // 申诉成功次数
	AppealTimes                int64   `db:"appeal_times" gorm:"column:appeal_times"`               // 申诉次数
	ApplicationTime            int64   `db:"application_time" gorm:"column:application_time"`       // 申请时间（毫秒时间戳）
	Avatar                     string  `db:"avatar" gorm:"column:avatar"`                           // 头像 URL
	Bank                       string  `db:"bank" gorm:"column:bank"`                               // 银行名称
	Branch                     string  `db:"branch" gorm:"column:branch"`                           // 银行支行
	CardNo                     string  `db:"card_no" gorm:"column:card_no"`                         // 银行卡号
	CertifiedBusinessApplyTime int64   `db:"certified_business_apply_time" gorm:"column:certified_business_apply_time"` // 认证商家申请时间（毫秒时间戳）
	CertifiedBusinessCheckTime int64   `db:"certified_business_check_time" gorm:"column:certified_business_check_time"` // 认证商家审核时间（毫秒时间戳）
	CertifiedBusinessStatus    int64   `db:"certified_business_status" gorm:"column:certified_business_status"` // 认证商家状态：0-未申请，1-审核中，2-已通过，3-已拒绝
	ChannelId                  int64   `db:"channel_id" gorm:"column:channel_id"`                   // 渠道 ID，用于渠道推广统计
	Email                      string  `db:"email" gorm:"column:email"`                             // 邮箱地址
	FirstLevel                 int64   `db:"first_level" gorm:"column:first_level"`                 // 一级下线数量，推广统计
	GoogleDate                 int64   `db:"google_date" gorm:"column:google_date"`                 // Google 验证器绑定时间（毫秒时间戳）
	GoogleKey                  string  `db:"google_key" gorm:"column:google_key"`                   // Google 验证器密钥
	GoogleState                int64   `db:"google_state" gorm:"column:google_state"`               // Google 验证器状态：0-未绑定，1-已绑定
	IdNumber                   string  `db:"id_number" gorm:"column:id_number"`                     // 身份证号
	InviterId                  int64   `db:"inviter_id" gorm:"column:inviter_id"`                   // 邀请人 ID，上级推广员
	IsChannel                  int64   `db:"is_channel" gorm:"column:is_channel"`                   // 是否为渠道商：0-否，1-是
	JyPassword                 string  `db:"jy_password" gorm:"column:jy_password"`                 // 交易密码（加密后），用于提现验证
	LastLoginTime              int64   `db:"last_login_time" gorm:"column:last_login_time"`         // 最后登录时间（毫秒时间戳）
	City                       string  `db:"city" gorm:"column:city"`                               // 城市
	Country                    string  `db:"country" gorm:"column:country"`                         // 国家
	District                   string  `db:"district" gorm:"column:district"`                       // 区/县
	Province                   string  `db:"province" gorm:"column:province"`                       // 省份
	LoginCount                 int64   `db:"login_count" gorm:"column:login_count"`                 // 登录次数
	LoginLock                  int64   `db:"login_lock" gorm:"column:login_lock"`                   // 登录锁定状态：0-未锁定，1-已锁定
	Margin                     string  `db:"margin" gorm:"column:margin"`                           // 保证金
	MemberLevel                int64   `db:"member_level" gorm:"column:member_level"`               // 会员等级：0-普通会员，1-实名会员，2-认证商家
	MobilePhone                string  `db:"mobile_phone" gorm:"column:mobile_phone"`               // 手机号，登录账号
	Password                   string  `db:"password" gorm:"column:password"`                       // 密码（加密后），登录密码
	PromotionCode              string  `db:"promotion_code" gorm:"column:promotion_code"`           // 推广码，用于下线注册
	PublishAdvertise           int64   `db:"publish_advertise" gorm:"column:publish_advertise"`     // 发布广告数
	RealName                   string  `db:"real_name" gorm:"column:real_name"`                     // 真实姓名
	RealNameStatus             int64   `db:"real_name_status" gorm:"column:real_name_status"`       // 实名认证状态：0-未认证，1-已认证
	RegistrationTime           int64   `db:"registration_time" gorm:"column:registration_time"`     // 注册时间（毫秒时间戳）
	Salt                       string  `db:"salt" gorm:"column:salt"`                               // 密码盐值，用于密码加密
	SecondLevel                int64   `db:"second_level" gorm:"column:second_level"`               // 二级下线数量，推广统计
	SignInAbility              int64   `db:"sign_in_ability" gorm:"column:sign_in_ability"`         // 签到能力
	Status                     int64   `db:"status" gorm:"column:status"`                           // 会员状态：0-正常，1-违规
	ThirdLevel                 int64   `db:"third_level" gorm:"column:third_level"`                 // 三级下线数量，推广统计
	Token                      string  `db:"token" gorm:"column:token"`                             // 登录 Token（已废弃，使用 JWT）
	TokenExpireTime            int64   `db:"token_expire_time" gorm:"column:token_expire_time"`     // Token 过期时间（已废弃）
	TransactionStatus          int64   `db:"transaction_status" gorm:"column:transaction_status"`   // 交易状态
	TransactionTime            int64   `db:"transaction_time" gorm:"column:transaction_time"`       // 交易时间（毫秒时间戳）
	Transactions               int64   `db:"transactions" gorm:"column:transactions"`               // 交易次数
	Username                   string  `db:"username" gorm:"column:username"`                       // 用户名，显示名称
	QrWeCodeUrl                string  `db:"qr_we_code_url" gorm:"column:qr_we_code_url"`           // 微信二维码 URL
	Wechat                     string  `db:"wechat" gorm:"column:wechat"`                           // 微信号
	Local                      string  `db:"local" gorm:"column:local"`                             // 地区
	Integration                int64   `db:"integration" gorm:"column:integration"`                 // 积分
	MemberGradeId              int64   `db:"member_grade_id" gorm:"column:member_grade_id"`         // 会员等级 ID
	KycStatus                  int64   `db:"kyc_status" gorm:"column:kyc_status"`                   // KYC 认证状态：0-未认证，1-认证中，2-已认证，3-认证失败
	GeneralizeTotal            int64   `db:"generalize_total" gorm:"column:generalize_total"`       // 推广总数
	InviterParentId            int64   `db:"inviter_parent_id" gorm:"column:inviter_parent_id"`     // 邀请人父 ID
	SuperPartner               string  `db:"super_partner" gorm:"column:super_partner"`             // 超级合伙人状态："0"-普通，"1"-超级合伙人，"2"-P 级
	KickFee                    float64 `db:"kick_fee" gorm:"column:kick_fee"`                       // 返佣手续费比例
	Power                      float64 `db:"power" gorm:"column:power"`                             // 算力
	TeamLevel                  int64   `db:"team_level" gorm:"column:team_level"`                   // 团队等级
	TeamPower                  float64 `db:"team_power" gorm:"column:team_power"`                   // 团队算力
	MemberLevelId              int64   `db:"member_level_id" gorm:"column:member_level_id"`         // 会员等级 ID
}

// --- 会员等级常量 ---

const (
	generalLevel       = iota // 普通会员：未实名认证
	realNameLevel              // 实名会员：已完成实名认证
	identificationLevel        // 认证商家：通过商家认证，可发布广告
)

// --- 合伙人状态常量 ---

const (
	normalPartner = "0" // 普通会员：无合伙人权益
	superPartner  = "1" // 超级合伙人：享受返佣权益
	pSuperPartner = "2" // P 级超级合伙人：更高级别的返佣权益
)

// --- 会员状态常量 ---

const (
	normalMemberStatus  = iota // 正常状态：会员功能正常
	illegalMemberStatus        // 违规状态：会员被限制部分功能
)

// MemberLevelText 返回会员等级文本描述
// 用于前端展示会员等级
func (m *Member) MemberLevelText() string {
	switch m.MemberLevel {
	case generalLevel:
		return "普通会员"
	case realNameLevel:
		return "实名"
	case identificationLevel:
		return "认证商家"
	default:
		return ""
	}
}

// MemberRate 返回会员费率等级
// 不同等级的合伙人享受不同的手续费费率
// 返回值：0-普通会员，1-超级合伙人，2-P 级超级合伙人
func (m *Member) MemberRate() int32 {
	switch m.SuperPartner {
	case superPartner:
		return 1
	case pSuperPartner:
		return 2
	default:
		return 0
	}
}

// FillSuperPartner 填充超级合伙人状态
// 根据传入的 partner 参数设置会员的合伙人状态和会员状态
//
// 业务规则：
//   - 空字符串：设置为普通会员，状态正常
//   - 非普通会员状态：设置为对应合伙人等级，状态标记为违规（需要审核）
//   - 普通会员状态：设置为普通会员，状态正常
func (m *Member) FillSuperPartner(partner string) {
	if partner == "" {
		m.SuperPartner = normalPartner
		m.Status = normalMemberStatus
		return
	}

	if partner != normalPartner {
		m.SuperPartner = partner
		m.Status = illegalMemberStatus
		return
	}

	m.SuperPartner = partner
	m.Status = normalMemberStatus
}

// NewMemberForRegister 构建旧版 MSCoin 注册流程所需的最小会员聚合。
//
// 为何在此处应用默认值而非处理器：
//   - 数据库 schema 较宽，历史上依赖零值填充
//   - 将默认值保留在模型中可保持传输层精简
//   - 所有注册入口点必须创建相同的初始会员状态
//
// 参数：
//   - now: 当前时间，用于设置注册时间
//   - phone: 手机号，作为登录账号
//   - username: 用户名，作为显示名称
//   - country: 国家，用于地区统计
//   - encodedPassword: 加密后的密码
//   - salt: 密码盐值
//   - partner: 合伙人状态
//   - promotion: 推广码
//
// 返回：初始化后的会员对象
func NewMemberForRegister(now time.Time, phone string, username string, country string, encodedPassword string, salt string, partner string, promotion string) *Member {
	member := &Member{
		AliNo:             "0",
		Avatar:            "https://mszlu.oss-cn-beijing.aliyuncs.com/mscoin/defaultavatar.png",
		Country:           country,
		LoginCount:        0,
		MemberLevel:       generalLevel,
		MobilePhone:       phone,
		Password:          encodedPassword,
		PromotionCode:     promotion,
		RegistrationTime:  now.UnixMilli(),
		Salt:              salt,
		TransactionStatus: 0,
		Username:          username,
	}
	member.FillSuperPartner(partner)
	return member
}
