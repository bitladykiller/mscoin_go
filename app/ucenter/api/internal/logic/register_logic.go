// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含用户注册相关的业务逻辑。
package logic

import (
	"context"
	"errors"
	"time"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// RegisterLogic 是用户注册业务逻辑处理器。
//
// 该结构体负责处理用户注册请求，支持多种注册方式。
type RegisterLogic struct {
	// ctx 是请求上下文。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewRegisterLogic 创建注册业务逻辑处理器实例。
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{ctx: ctx, svcCtx: svcCtx}
}

// Register 执行用户注册业务逻辑。
//
// 注册流程：
//  1. 验证请求参数完整性（验证码必须存在）
//  2. 调用 ucenter-rpc RegisterClient 执行注册
//  3. RPC 服务验证短信验证码、创建会员账号
//
// 支持的注册方式：
//   - 手机号注册：需要手机号和短信验证码
//   - 邀请注册：可携带邀请码建立邀请关系
//
// RPC 调用：
//   - RegisterClient.RegisterByPhone -> ucenter-rpc
//   - ucenter-rpc 负责：
//       - 验证验证码
//       - 验证短信验证码
//       - 检查用户名/手机号是否已存在
//       - 创建会员账号
//       - 建立邀请关系（如有邀请码）
//       - 初始化用户钱包
//
// 参数：
//   - req：注册请求，包含用户名、密码、手机号、验证码、邀请码等
//
// 返回：
//   - *types.Response：注册成功响应
//   - error：注册失败时的错误信息
func (l *RegisterLogic) Register(req *types.Request) (*types.Response, error) {
	// 验证参数完整性
	// 验证码是防止机器人注册的必要条件
	if req == nil || req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}

	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 调用 ucenter-rpc RegisterClient 执行手机号注册
	_, err := l.svcCtx.RegisterClient.RegisterByPhone(ctx, &registerpb.RegReq{
		Username: req.Username,
		Password: req.Password,
		Captcha: &registerpb.CaptchaReq{
			Server: req.Captcha.Server,
			Token:  req.Captcha.Token,
		},
		Phone:        req.Phone,
		Promotion:    req.Promotion,
		Code:         req.Code,
		Country:      req.Country,
		SuperPartner: req.SuperPartner,
		Ip:           req.IP,
	})
	if err != nil {
		return nil, err
	}

	return &types.Response{}, nil
}
