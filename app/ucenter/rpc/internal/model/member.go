package model

import "time"

// Member 会员模型
// 存储会员的基本信息、认证状态、钱包关联等数据
type Member struct {
	Id                         int64   `db:"id" gorm:"column:id"`                                   // 会员 ID
	AliNo                      string  `db:"ali_no" gorm:"column:ali_no"`                           // 支付宝账号
	QrCodeUrl                  string  `db:"qr_code_url" gorm:"column:qr_code_url"`                 // 二维码 URL
	AppealSuccessTimes         int64   `db:"appeal_success_times" gorm:"column:appeal_success_times"` // 申诉成功次数
	AppealTimes                int64   `db:"appeal_times" gorm:"column:appeal_times"`               // 申诉次数
	ApplicationTime            int64   `db:"application_time" gorm:"column:application_time"`       // 申请时间
	Avatar                     string  `db:"avatar" gorm:"column:avatar"`                           // 头像 URL
	Bank                       string  `db:"bank" gorm:"column:bank"`                               // 银行名称
	Branch                     string  `db:"branch" gorm:"column:branch"`                           // 银行支行
	CardNo                     string  `db:"card_no" gorm:"column:card_no"`                         // 银行卡号
	CertifiedBusinessApplyTime int64   `db:"certified_business_apply_time" gorm:"column:certified_business_apply_time"` // 认证商家申请时间
	CertifiedBusinessCheckTime int64   `db:"certified_business_check_time" gorm:"column:certified_business_check_time"` // 认证商家审核时间
	CertifiedBusinessStatus    int64   `db:"certified_business_status" gorm:"column:certified_business_status"` // 认证商家状态
	ChannelId                  int64   `db:"channel_id" gorm:"column:channel_id"`                   // 渠道 ID
	Email                      string  `db:"email" gorm:"column:email"`                             // 邮箱
	FirstLevel                 int64   `db:"first_level" gorm:"column:first_level"`                 // 一级下线数量
	GoogleDate                 int64   `db:"google_date" gorm:"column:google_date"`                 // Google 验证器绑定时间
	GoogleKey                  string  `db:"google_key" gorm:"column:google_key"`                   // Google 验证器密钥
	GoogleState                int64   `db:"google_state" gorm:"column:google_state"`               // Google 验证器状态
	IdNumber                   string  `db:"id_number" gorm:"column:id_number"`                     // 身份证号
	InviterId                  int64   `db:"inviter_id" gorm:"column:inviter_id"`                   // 邀请人 ID
	IsChannel                  int64   `db:"is_channel" gorm:"column:is_channel"`                   // 是否为渠道商
	JyPassword                 string  `db:"jy_password" gorm:"column:jy_password"`                 // 交易密码
	LastLoginTime              int64   `db:"last_login_time" gorm:"column:last_login_time"`         // 最后登录时间
	City                       string  `db:"city" gorm:"column:city"`                               // 城市
	Country                    string  `db:"country" gorm:"column:country"`                         // 国家
	District                   string  `db:"district" gorm:"column:district"`                       // 区/县
	Province                   string  `db:"province" gorm:"column:province"`                       // 省份
	LoginCount                 int64   `db:"login_count" gorm:"column:login_count"`                 // 登录次数
	LoginLock                  int64   `db:"login_lock" gorm:"column:login_lock"`                   // 登录锁定状态
	Margin                     string  `db:"margin" gorm:"column:margin"`                           // 保证金
	MemberLevel                int64   `db:"member_level" gorm:"column:member_level"`               // 会员等级
	MobilePhone                string  `db:"mobile_phone" gorm:"column:mobile_phone"`               // 手机号
	Password                   string  `db:"password" gorm:"column:password"`                       // 密码（加密后）
	PromotionCode              string  `db:"promotion_code" gorm:"column:promotion_code"`           // 推广码
	PublishAdvertise           int64   `db:"publish_advertise" gorm:"column:publish_advertise"`     // 发布广告数
	RealName                   string  `db:"real_name" gorm:"column:real_name"`                     // 真实姓名
	RealNameStatus             int64   `db:"real_name_status" gorm:"column:real_name_status"`       // 实名认证状态
	RegistrationTime           int64   `db:"registration_time" gorm:"column:registration_time"`     // 注册时间
	Salt                       string  `db:"salt" gorm:"column:salt"`                               // 密码盐值
	SecondLevel                int64   `db:"second_level" gorm:"column:second_level"`               // 二级下线数量
	SignInAbility              int64   `db:"sign_in_ability" gorm:"column:sign_in_ability"`         // 签到能力
	Status                     int64   `db:"status" gorm:"column:status"`                           // 会员状态
	ThirdLevel                 int64   `db:"third_level" gorm:"column:third_level"`                 // 三级下线数量
	Token                      string  `db:"token" gorm:"column:token"`                             // 登录 Token
	TokenExpireTime            int64   `db:"token_expire_time" gorm:"column:token_expire_time"`     // Token 过期时间
	TransactionStatus          int64   `db:"transaction_status" gorm:"column:transaction_status"`   // 交易状态
	TransactionTime            int64   `db:"transaction_time" gorm:"column:transaction_time"`       // 交易时间
	Transactions               int64   `db:"transactions" gorm:"column:transactions"`               // 交易次数
	Username                   string  `db:"username" gorm:"column:username"`                       // 用户名
	QrWeCodeUrl                string  `db:"qr_we_code_url" gorm:"column:qr_we_code_url"`           // 微信二维码 URL
	Wechat                     string  `db:"wechat" gorm:"column:wechat"`                           // 微信号
	Local                      string  `db:"local" gorm:"column:local"`                             // 地区
	Integration                int64   `db:"integration" gorm:"column:integration"`                 // 积分
	MemberGradeId              int64   `db:"member_grade_id" gorm:"column:member_grade_id"`         // 会员等级 ID
	KycStatus                  int64   `db:"kyc_status" gorm:"column:kyc_status"`                   // KYC 认证状态
	GeneralizeTotal            int64   `db:"generalize_total" gorm:"column:generalize_total"`       // 推广总数
	InviterParentId            int64   `db:"inviter_parent_id" gorm:"column:inviter_parent_id"`     // 邀请人父 ID
	SuperPartner               string  `db:"super_partner" gorm:"column:super_partner"`             // 超级合伙人状态
	KickFee                    float64 `db:"kick_fee" gorm:"column:kick_fee"`                       // 返佣手续费
	Power                      float64 `db:"power" gorm:"column:power"`                             // 算力
	TeamLevel                  int64   `db:"team_level" gorm:"column:team_level"`                   // 团队等级
	TeamPower                  float64 `db:"team_power" gorm:"column:team_power"`                   // 团队算力
	MemberLevelId              int64   `db:"member_level_id" gorm:"column:member_level_id"`         // 会员等级 ID
}

// 会员等级常量
const (
	generalLevel       = iota // 普通会员
	realNameLevel              // 实名会员
	identificationLevel        // 认证商家
)

// 合伙人状态常量
const (
	normalPartner = "0" // 普通会员
	superPartner  = "1" // 超级合伙人
	pSuperPartner = "2" // P 级超级合伙人
)

// 会员状态常量
const (
	normalMemberStatus  = iota // 正常状态
	illegalMemberStatus        // 违规状态
)

// MemberLevelText 返回会员等级文本描述
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
// - 数据库 schema 较宽，历史上依赖零值填充
// - 将默认值保留在模型中可保持传输层精简
// - 所有注册入口点必须创建相同的初始会员状态
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
