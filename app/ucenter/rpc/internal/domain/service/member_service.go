// Package service 定义会员领域服务。
//
// MemberService 是会员管理的核心领域服务，负责：
//   - 会员登录：验证码校验、密码验证、Token 生成
//   - 会员注册：验证码校验、密码加密、会员创建
//   - 验证码发送：生成并缓存验证码
//   - 会员信息查询：按 ID 查询会员详情
//
// 设计原则：
//   - 单一职责：只处理会员相关的业务逻辑
//   - 依赖注入：通过构造函数注入仓储和缓存
//   - 接口隔离：不依赖具体实现，便于测试
//
// 安全设计：
//   - 密码加密存储（加盐哈希）
//   - 验证码校验防止机器人
//   - JWT Token 用于无状态认证
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
// 格式：REGISTER::{phone}
// 用于缓存发送给指定手机号的验证码
const registerCacheKey = "REGISTER::"

// MemberService 会员服务
// 负责会员登录、注册、信息查询等核心业务逻辑
//
// 依赖：
//   - repo: 会员仓储，用于数据持久化
//   - captchaService: 验证码服务，用于人机验证
//   - cache: Redis 缓存，用于存储验证码
//   - cfg: 服务配置，包含 JWT 密钥等
type MemberService struct {
	repo           *repository.MemberRepository // 会员仓储
	captchaService *CaptchaService             // 验证码服务
	cache          *redisx.Client               // Redis 缓存客户端
	cfg            config.Config                // 服务配置
}

// NewMemberService 创建会员服务实例
// 参数通过依赖注入，便于单元测试时 Mock
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
//
// 登录流程：
//  1. 验证人机验证码（Captcha）
//  2. 查询会员信息
//  3. 验证密码（加盐哈希比对）
//  4. 生成 JWT Token
//  5. 异步更新登录次数
//
// 参数：
//   - ctx: 请求上下文
//   - req: 登录请求，包含用户名、密码、验证码等
//
// 返回：
//   - LoginRes: 登录响应，包含 Token、会员信息等
//   - error: 错误信息（验证码失败、用户不存在、密码错误等）
func (s *MemberService) Login(ctx context.Context, req *loginpb.LoginReq) (*loginpb.LoginRes, error) {
	// 验证验证码
	// 防止机器人暴力破解密码
	if req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}
	if !s.captchaService.Verify(req.Captcha.Server, s.cfg.Captcha.Vid, s.cfg.Captcha.Key, req.Captcha.Token, 2, req.Ip) {
		return nil, errors.New("captcha verification failed")
	}

	// 查找会员
	// 使用手机号作为登录账号
	member, err := s.repo.FindByPhone(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("user not registered")
	}

	// 验证密码
	// 使用加盐哈希验证，防止彩虹表攻击
	if !passwordx.Verify(req.Password, member.Salt, member.Password) {
		return nil, errors.New("wrong password")
	}

	// 生成 JWT Token
	// Token 包含会员 ID，过期时间由配置决定
	token, err := auth.GenerateUserToken(s.cfg.JWT.AccessSecret, time.Now(), s.cfg.JWT.AccessExpire, member.Id)
	if err != nil {
		return nil, errors.New("generate token failed")
	}

	// 异步更新登录次数
	// 不阻塞登录响应，提高用户体验
	// 使用新的 context.Background() 避免请求取消影响
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
// 用于会员详情页面、身份验证等场景
//
// 参数：
//   - ctx: 请求上下文
//   - memberID: 会员 ID
//
// 返回：
//   - MemberInfo: 会员信息（完整字段）
//   - error: 错误信息（会员不存在等）
func (s *MemberService) FindByID(ctx context.Context, memberID int64) (*memberpb.MemberInfo, error) {
	member, err := s.repo.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}

	// 转换为 protobuf 响应
	// 包含会员的所有字段，用于前端展示
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
// 完成验证码校验、密码加密、会员创建
//
// 注册流程：
//  1. 验证人机验证码（Captcha）
//  2. 验证短信验证码（Redis 缓存比对）
//  3. 检查手机号是否已注册
//  4. 密码加密（加盐哈希）
//  5. 创建会员记录
//
// 参数：
//   - ctx: 请求上下文
//   - req: 注册请求，包含手机号、用户名、密码、验证码等
//
// 返回：
//   - RegRes: 注册响应（空响应）
//   - error: 错误信息（验证码失败、手机号已注册等）
func (s *MemberService) RegisterByPhone(ctx context.Context, req *registerpb.RegReq) (*registerpb.RegRes, error) {
	// 验证验证码
	// 防止机器人批量注册
	if req == nil || req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}
	if !s.captchaService.Verify(req.Captcha.Server, s.cfg.Captcha.Vid, s.cfg.Captcha.Key, req.Captcha.Token, 2, req.Ip) {
		return nil, errors.New("captcha verification failed")
	}

	// 验证短信验证码
	// 从 Redis 获取缓存的验证码进行比对
	var cachedCode string
	if err := s.cache.GetCtx(ctx, registerCacheKey+req.Phone, &cachedCode); err != nil {
		return nil, errors.New("verification code unavailable")
	}
	if cachedCode != req.Code {
		return nil, errors.New("verification code mismatch")
	}

	// 检查手机号是否已注册
	// 防止重复注册
	member, err := s.repo.FindByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if member != nil {
		return nil, errors.New("phone already registered")
	}

	// 编码密码并创建会员
	// 密码加密流程：
	//  1. 生成随机盐值
	//  2. 密码 + 盐值 -> 哈希
	salt, encodedPassword := passwordx.Encode(req.Password)
	newMember := model.NewMemberForRegister(time.Now(), req.Phone, req.Username, req.Country, encodedPassword, salt, req.SuperPartner, req.Promotion)
	if err := s.repo.Save(ctx, newMember); err != nil {
		return nil, errors.New("register failed")
	}

	return &registerpb.RegRes{}, nil
}

// SendRegisterCode 发送注册验证码
// 生成验证码并缓存到 Redis
//
// 验证码规则：
//   - 长度：4 位数字
//   - 有效期：15 分钟
//   - 缓存键：REGISTER::{phone}
//
// 注意：实际发送短信由外部服务完成，本方法只负责生成和缓存
//
// 参数：
//   - ctx: 请求上下文
//   - req: 验证码请求，包含手机号
//
// 返回：
//   - NoRes: 空响应
//   - error: 错误信息（手机号为空等）
func (s *MemberService) SendRegisterCode(ctx context.Context, req *registerpb.CodeReq) (*registerpb.NoRes, error) {
	if req == nil || req.Phone == "" {
		return nil, errors.New("phone is required")
	}

	// 生成 4 位数字验证码
	// 使用加密安全的随机数生成器
	code, err := generateNumericCode(4)
	if err != nil {
		return nil, errors.New("generate verification code failed")
	}

	// 缓存验证码，有效期 15 分钟
	// 后续注册时需要验证此验证码
	if err := s.cache.SetWithExpireCtx(ctx, registerCacheKey+req.Phone, code, 15*time.Minute); err != nil {
		return nil, errors.New("send verification code failed")
	}

	return &registerpb.NoRes{}, nil
}

// generateNumericCode 生成指定长度的数字验证码
// 使用加密安全的随机数生成器
//
// 参数：
//   - length: 验证码长度
//
// 返回：
//   - string: 数字验证码
//   - error: 错误信息（长度无效等）
func generateNumericCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}

	buffer := make([]byte, length)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	// 将随机字节转换为数字字符
	// 每个字节模 10 得到 0-9 的数字
	for index, value := range buffer {
		buffer[index] = '0' + (value % 10)
	}
	return string(buffer), nil
}
