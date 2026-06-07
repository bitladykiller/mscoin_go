package model

import "time"

type Member struct {
	Id                         int64   `db:"id" gorm:"column:id"`
	AliNo                      string  `db:"ali_no" gorm:"column:ali_no"`
	QrCodeUrl                  string  `db:"qr_code_url" gorm:"column:qr_code_url"`
	AppealSuccessTimes         int64   `db:"appeal_success_times" gorm:"column:appeal_success_times"`
	AppealTimes                int64   `db:"appeal_times" gorm:"column:appeal_times"`
	ApplicationTime            int64   `db:"application_time" gorm:"column:application_time"`
	Avatar                     string  `db:"avatar" gorm:"column:avatar"`
	Bank                       string  `db:"bank" gorm:"column:bank"`
	Branch                     string  `db:"branch" gorm:"column:branch"`
	CardNo                     string  `db:"card_no" gorm:"column:card_no"`
	CertifiedBusinessApplyTime int64   `db:"certified_business_apply_time" gorm:"column:certified_business_apply_time"`
	CertifiedBusinessCheckTime int64   `db:"certified_business_check_time" gorm:"column:certified_business_check_time"`
	CertifiedBusinessStatus    int64   `db:"certified_business_status" gorm:"column:certified_business_status"`
	ChannelId                  int64   `db:"channel_id" gorm:"column:channel_id"`
	Email                      string  `db:"email" gorm:"column:email"`
	FirstLevel                 int64   `db:"first_level" gorm:"column:first_level"`
	GoogleDate                 int64   `db:"google_date" gorm:"column:google_date"`
	GoogleKey                  string  `db:"google_key" gorm:"column:google_key"`
	GoogleState                int64   `db:"google_state" gorm:"column:google_state"`
	IdNumber                   string  `db:"id_number" gorm:"column:id_number"`
	InviterId                  int64   `db:"inviter_id" gorm:"column:inviter_id"`
	IsChannel                  int64   `db:"is_channel" gorm:"column:is_channel"`
	JyPassword                 string  `db:"jy_password" gorm:"column:jy_password"`
	LastLoginTime              int64   `db:"last_login_time" gorm:"column:last_login_time"`
	City                       string  `db:"city" gorm:"column:city"`
	Country                    string  `db:"country" gorm:"column:country"`
	District                   string  `db:"district" gorm:"column:district"`
	Province                   string  `db:"province" gorm:"column:province"`
	LoginCount                 int64   `db:"login_count" gorm:"column:login_count"`
	LoginLock                  int64   `db:"login_lock" gorm:"column:login_lock"`
	Margin                     string  `db:"margin" gorm:"column:margin"`
	MemberLevel                int64   `db:"member_level" gorm:"column:member_level"`
	MobilePhone                string  `db:"mobile_phone" gorm:"column:mobile_phone"`
	Password                   string  `db:"password" gorm:"column:password"`
	PromotionCode              string  `db:"promotion_code" gorm:"column:promotion_code"`
	PublishAdvertise           int64   `db:"publish_advertise" gorm:"column:publish_advertise"`
	RealName                   string  `db:"real_name" gorm:"column:real_name"`
	RealNameStatus             int64   `db:"real_name_status" gorm:"column:real_name_status"`
	RegistrationTime           int64   `db:"registration_time" gorm:"column:registration_time"`
	Salt                       string  `db:"salt" gorm:"column:salt"`
	SecondLevel                int64   `db:"second_level" gorm:"column:second_level"`
	SignInAbility              int64   `db:"sign_in_ability" gorm:"column:sign_in_ability"`
	Status                     int64   `db:"status" gorm:"column:status"`
	ThirdLevel                 int64   `db:"third_level" gorm:"column:third_level"`
	Token                      string  `db:"token" gorm:"column:token"`
	TokenExpireTime            int64   `db:"token_expire_time" gorm:"column:token_expire_time"`
	TransactionStatus          int64   `db:"transaction_status" gorm:"column:transaction_status"`
	TransactionTime            int64   `db:"transaction_time" gorm:"column:transaction_time"`
	Transactions               int64   `db:"transactions" gorm:"column:transactions"`
	Username                   string  `db:"username" gorm:"column:username"`
	QrWeCodeUrl                string  `db:"qr_we_code_url" gorm:"column:qr_we_code_url"`
	Wechat                     string  `db:"wechat" gorm:"column:wechat"`
	Local                      string  `db:"local" gorm:"column:local"`
	Integration                int64   `db:"integration" gorm:"column:integration"`
	MemberGradeId              int64   `db:"member_grade_id" gorm:"column:member_grade_id"`
	KycStatus                  int64   `db:"kyc_status" gorm:"column:kyc_status"`
	GeneralizeTotal            int64   `db:"generalize_total" gorm:"column:generalize_total"`
	InviterParentId            int64   `db:"inviter_parent_id" gorm:"column:inviter_parent_id"`
	SuperPartner               string  `db:"super_partner" gorm:"column:super_partner"`
	KickFee                    float64 `db:"kick_fee" gorm:"column:kick_fee"`
	Power                      float64 `db:"power" gorm:"column:power"`
	TeamLevel                  int64   `db:"team_level" gorm:"column:team_level"`
	TeamPower                  float64 `db:"team_power" gorm:"column:team_power"`
	MemberLevelId              int64   `db:"member_level_id" gorm:"column:member_level_id"`
}

const (
	generalLevel = iota
	realNameLevel
	identificationLevel
)

const (
	normalPartner = "0"
	superPartner  = "1"
	pSuperPartner = "2"
)

const (
	normalMemberStatus = iota
	illegalMemberStatus
)

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

// NewMemberForRegister constructs the minimum member aggregate required by the
// legacy MSCoin registration flow.
//
// Why defaults are applied here instead of the handler:
// - the database schema is wide and historically relied on zero-value filling
// - keeping defaults in the model keeps transport layers thin
// - all register entry points must create the same initial member state
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
