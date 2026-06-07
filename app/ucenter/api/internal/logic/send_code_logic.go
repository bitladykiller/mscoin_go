// Package logic 提供 ucenter-api 服务的业务逻辑处理。
//
// 该文件包含发送短信验证码相关的业务逻辑。
package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// SendCodeLogic 是发送验证码业务逻辑处理器。
//
// 该结构体负责处理发送短信验证码请求。
type SendCodeLogic struct {
	// ctx 是请求上下文。
	ctx    context.Context

	// svcCtx 是服务上下文，提供 RPC 客户端访问能力。
	svcCtx *svc.ServiceContext
}

// NewSendCodeLogic 创建发送验证码业务逻辑处理器实例。
func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendCode 执行发送短信验证码业务逻辑。
//
// 发送流程：
//  1. 调用 ucenter-rpc RegisterClient 发送验证码
//  2. RPC 服务验证手机号格式、调用短信服务
//
// RPC 调用：
//   - RegisterClient.SendCode -> ucenter-rpc
//   - ucenter-rpc 负责：
//       - 验证手机号格式
//       - 检查发送频率限制
//       - 生成验证码
//       - 调用短信服务商 API 发送验证码
//       - 缓存验证码供后续验证
//
// 参数：
//   - req：发送验证码请求，包含手机号和国家代码
//
// 返回：
//   - *types.CodeResponse：发送成功响应
//   - error：发送失败时的错误信息
func (l *SendCodeLogic) SendCode(req *types.CodeRequest) (*types.CodeResponse, error) {
	// 设置 RPC 调用超时
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	// 调用 ucenter-rpc RegisterClient 发送短信验证码
	_, err := l.svcCtx.RegisterClient.SendCode(ctx, &registerpb.CodeReq{
		Phone:   req.Phone,
		Country: req.Country,
	})
	if err != nil {
		return nil, err
	}

	return &types.CodeResponse{}, nil
}
