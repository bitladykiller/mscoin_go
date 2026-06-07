package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"mscoin_go/app/ucenter/rpc/internal/config"
	"mscoin_go/app/ucenter/rpc/internal/model"
	"mscoin_go/app/ucenter/rpc/internal/repository"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
	"mscoin_go/pkg/auth"
	"mscoin_go/pkg/cache/redisx"
	"mscoin_go/pkg/passwordx"
)

// registerCacheKey 注册验证码缓存键前缀
const registerCacheKey = "REGISTER::"

// MemberService 会员服务
// 负责会员登录、注册、信息查询等核心业务逻辑
type MemberService struct {
	repo           *repository.MemberRepository // 会员仓储
	captchaService *CaptchaService             // 验证码服务
	cache          *redisx.Client               // Redis 缓存客户端
	cfg            config.Config                // 服务配置
}

// NewMemberService 创建会员服务实例
func NewMemberService(repo *repository.MemberRepository, captchaService *CaptchaService, cache *redisx.Client, cfg config.Config) *MemberService {
	return &MemberService{
		repo:           repo,
		captchaService: captchaService,
		cache:          cache,
		cfg:            cfg,
	}
}

// Login 处理会员登录
// 验证验证码、密码，生成 JWT Token
func (s *MemberService) Login(ctx context.Context, req *loginpb.LoginReq) (*loginpb.LoginRes, error) {
	// 验证验证码
	if req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}
	if !s.captchaService.Verify(req.Captcha.Server, s.cfg.Captcha.Vid, s.cfg.Captcha.Key, req.Captcha.Token, 2, req.Ip) {
		return nil, errors.New("captcha verification failed")
	}

	// 查找会员
	member, err := s.repo.FindByPhone(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("user not registered")
	}

	// 验证密码
	if !passwordx.Verify(req.Password, member.Salt, member.Password) {
		return nil, errors.New("wrong password")
	}

	// 生成 JWT Token
	token, err := auth.GenerateUserToken(s.cfg.JWT.AccessSecret, time.Now(), s.cfg.JWT.AccessExpire, member.Id)
	if err != nil {
		return nil, errors.New("generate token failed")
	}

	// 异步更新登录次数
	go func() {
		_ = s.repo.UpdateLoginCount(context.Background(), member.Id, 1)
	}()

	return &loginpb.LoginRes{
		Username:      member.Username,
		Token:         token,
		MemberLevel:   member.MemberLevelText(),
		RealName:      member.RealName,
		Country:       member.Country,
		Avatar:        member.Avatar,
		PromotionCode: member.PromotionCode,
		Id:            member.Id,
		LoginCount:    int32(member.LoginCount + 1),
		SuperPartner:  member.SuperPartner,
		MemberRate:    member.MemberRate(),
	}, nil
}

// FindByID 根据会员 ID 查询会员信息
func (s *MemberService) FindByID(ctx context.Context, memberID int64) (*memberpb.MemberInfo, error) {
	member, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}

	return &memberpb.MemberInfo{
		Id:                         member.Id,
		AliNo:                      member.AliNo,
		QrCodeUrl:                  member.QrCodeUrl,
		AppealSuccessTimes:         int32(member.AppealSuccessTimes),
		AppealTimes:                int32(member.AppealTimes),
		ApplicationTime:            member.ApplicationTime,
		Avatar:                     member.Avatar,
		Bank:                       member.Bank,
		Branch:                     member.Branch,
		CardNo:                     member.CardNo,
		CertifiedBusinessApplyTime: member.CertifiedBusinessApplyTime,
		CertifiedBusinessCheckTime: member.CertifiedBusinessCheckTime,
		CertifiedBusinessStatus:    int32(member.CertifiedBusinessStatus),
		ChannelId:                  int32(member.ChannelId),
		Email:                      member.Email,
		FirstLevel:                 int32(member.FirstLevel),
		GoogleDate:                 member.GoogleDate,
		GoogleKey:                  member.GoogleKey,
		GoogleState:                int32(member.GoogleState),
		IdNumber:                   member.IdNumber,
		InviterId:                  member.InviterId,
		IsChannel:                  int32(member.IsChannel),
		JyPassword:                 member.JyPassword,
		LastLoginTime:              member.LastLoginTime,
		City:                       member.City,
		Country:                    member.Country,
		District:                   member.District,
		Province:                   member.Province,
		LoginCount:                 int32(member.LoginCount),
		LoginLock:                  int32(member.LoginLock),
		Margin:                     member.Margin,
		MemberLevel:                int32(member.MemberLevel),
		MobilePhone:                member.MobilePhone,
		Password:                   member.Password,
		PromotionCode:              member.PromotionCode,
		PublishAdvertise:           int32(member.PublishAdvertise),
		RealName:                   member.RealName,
		RealNameStatus:             int32(member.RealNameStatus),
		RegistrationTime:           member.RegistrationTime,
		Salt:                       member.Salt,
		SecondLevel:                int32(member.SecondLevel),
		SignInAbility:              int32(member.SignInAbility),
		Status:                     int32(member.Status),
		ThirdLevel:                 int32(member.ThirdLevel),
		Token:                      member.Token,
		TokenExpireTime:            member.TokenExpireTime,
		TransactionStatus:          int32(member.TransactionStatus),
		TransactionTime:            member.TransactionTime,
		Transactions:               int32(member.Transactions),
		Username:                   member.Username,
		QrWeCodeUrl:                member.QrWeCodeUrl,
		Wechat:                     member.Wechat,
		Local:                      member.Local,
		Integration:                member.Integration,
		MemberGradeId:              member.MemberGradeId,
		KycStatus:                  int32(member.KycStatus),
		GeneralizeTotal:            member.GeneralizeTotal,
		InviterParentId:            member.InviterParentId,
		SuperPartner:               member.SuperPartner,
		KickFee:                    member.KickFee,
		Power:                      member.Power,
		TeamLevel:                  int32(member.TeamLevel),
		TeamPower:                  member.TeamPower,
		MemberLevelId:              member.MemberLevelId,
	}, nil
}

// RegisterByPhone 通过手机号注册会员
func (s *MemberService) RegisterByPhone(ctx context.Context, req *registerpb.RegReq) (*registerpb.RegRes, error) {
	// 验证验证码
	if req == nil || req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}
	if !s.captchaService.Verify(req.Captcha.Server, s.cfg.Captcha.Vid, s.cfg.Captcha.Key, req.Captcha.Token, 2, req.Ip) {
		return nil, errors.New("captcha verification failed")
	}

	// 验证短信验证码
	var cachedCode string
	if err := s.cache.GetCtx(ctx, registerCacheKey+req.Phone, &cachedCode); err != nil {
		return nil, errors.New("verification code unavailable")
	}
	if cachedCode != req.Code {
		return nil, errors.New("verification code mismatch")
	}

	// 检查手机号是否已注册
	member, err := s.repo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if member != nil {
		return nil, errors.New("phone already registered")
	}

	// 编码密码并创建会员
	salt, encodedPassword := passwordx.Encode(req.Password)
	newMember := model.NewMemberForRegister(time.Now(), req.Phone, req.Username, req.Country, encodedPassword, salt, req.SuperPartner, req.Promotion)
	if err := s.repo.Save(ctx, newMember); err != nil {
		return nil, errors.New("register failed")
	}

	return &registerpb.RegRes{}, nil
}

// SendRegisterCode 发送注册验证码
func (s *MemberService) SendRegisterCode(ctx context.Context, req *registerpb.CodeReq) (*registerpb.NoRes, error) {
	if req == nil || req.Phone == "" {
		return nil, errors.New("phone is required")
	}

	// 生成 4 位数字验证码
	code, err := generateNumericCode(4)
	if err != nil {
		return nil, errors.New("generate verification code failed")
	}

	// 缓存验证码，有效期 15 分钟
	if err := s.cache.SetWithExpireCtx(ctx, registerCacheKey+req.Phone, code, 15*time.Minute); err != nil {
		return nil, errors.New("send verification code failed")
	}

	return &registerpb.NoRes{}, nil
}

// generateNumericCode 生成指定长度的数字验证码
func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	for index, value := range buffer {
		buffer[index] = '0' + (value % 10)
	}
	return string(buffer), nil
}
