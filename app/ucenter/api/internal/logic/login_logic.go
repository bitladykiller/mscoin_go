// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该包包含所有 HTTP 接口的业务逻辑实现，负责：
//   - 接收 handler 层传递的请求参数
//   - 调用 RPC 服务完成业务操作
//   - 处理数据转换和组装
//   - 返回结果给 handler 层
//
// 架构设计：
//   - 每个接口对应一个 Logic 结构体
//   - Logic 持有 context 和 ServiceContext
//   - 通过 ServiceContext 访问 RPC 客户端
package logic

import (
	"context"
	"errors"
	"time"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
)

// LoginLogic 是用户登录业务逻辑处理器。
//
// 该结构体负责处理用户登录请求，验证用户身份并返回 JWT Token。
type LoginLogic struct {
	// ctx 是请求上下文，用于传递用户身份和取消信号。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 创建登录业务逻辑处理器实例。
//
// 参数：
//   - ctx：请求上下文
//   - svcCtx：服务上下文，包含 RPC 客户端
//
// 返回：
//   - *LoginLogic：登录逻辑处理器
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx: ctx, svcCtx: svcCtx}
}

// Login 执行用户登录业务逻辑。
//
// 登录流程：
//  1. 验证请求参数完整性（验证码必须存在）
//  2. 调用 ucenter-rpc LoginClient 验证用户名密码
//  3. RPC 服务验证通过后返回 JWT Token 和用户信息
//  4. 转换 RPC 响应为 API 响应格式
//
// 为什么在这里也做防护：
//   - HTTP handler 已经为真实请求验证了验证码的存在性
//   - 但 logic 层仍可能被测试或未来的内部调用者直接复用
//   - 在 logic 层保持不变性检查可以防止绕过传输层适配器导致的空指针异常
//
// RPC 调用：
//   - LoginClient.Login -> ucenter-rpc
//   - ucenter-rpc 负责：
//       - 验证验证码
//       - 验证用户名密码
//       - 生成 JWT Token
//       - 记录登录日志
//       - 更新登录统计
//
// 参数：
//   - req：登录请求，包含用户名、密码、验证码、IP
//
// 返回：
//   - *types.LoginRes：登录响应，包含 Token 和用户信息
//   - error：登录失败时的错误信息
func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginRes, error) {
	// 为什么在这里也做防护：
	// - HTTP handler 已经为真实请求验证了验证码的存在性
	// - 但 logic 层仍可能被测试或未来的内部调用者直接复用
	// - 在 logic 层保持不变性检查可以防止绕过传输层适配器导致的空指针异常
	if req == nil || req.Captcha == nil {
		return nil, errors.New("captcha verification failed")
	}

	// 设置 RPC 调用超时，防止长时间阻塞
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 调用 ucenter-rpc LoginClient 执行登录验证
	// RPC 服务会验证验证码、用户名密码，并生成 JWT Token
	payload, err := l.svcCtx.LoginClient.Login(ctx, &loginpb.LoginReq{
		Username: req.Username,
		Password: req.Password,
		Captcha: &loginpb.CaptchaReq{
			Server: req.Captcha.Server,
			Token:  req.Captcha.Token,
		},
		Ip: req.IP,
	})
	if err != nil {
		return nil, err
	}

	// 转换 RPC 响应为 API 响应格式
	return &types.LoginRes{
		Username:      payload.Username,
		Token:         payload.Token,
		MemberLevel:   payload.MemberLevel,
		RealName:      payload.RealName,
		Country:       payload.Country,
		Avatar:        payload.Avatar,
		PromotionCode: payload.PromotionCode,
		Id:            payload.Id,
		LoginCount:    int(payload.LoginCount),
		SuperPartner:  payload.SuperPartner,
		MemberRate:    int(payload.MemberRate),
	}, nil
}
