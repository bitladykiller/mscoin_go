// Package server 定义注册 RPC 服务端。
package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// RegisterServer 注册 RPC 服务端
// 处理会员注册相关的 RPC 请求
//
// 实现 registerpb.RegisterServer 接口
// 提供以下方法：
//   - RegisterByPhone: 处理手机号注册请求
//   - SendCode: 发送注册验证码
type RegisterServer struct {
	svcCtx *svc.ServiceContext           // 服务上下文
	registerpb.UnimplementedRegisterServer // 未实现方法的默认实现
}

// NewRegisterServer 创建注册服务端实例
func NewRegisterServer(svcCtx *svc.ServiceContext) *RegisterServer {
	return &RegisterServer{svcCtx: svcCtx}
}

// RegisterByPhone 处理手机号注册请求
// 接收 gRPC 请求并转发给 RegisterByPhoneLogic 处理
//
// 参数：
//   - ctx: 请求上下文
//   - in: 注册请求，包含手机号、用户名、密码、验证码等
//
// 返回：
//   - RegRes: 注册响应（空响应）
//   - error: 错误信息
func (s *RegisterServer) RegisterByPhone(ctx context.Context, in *registerpb.RegReq) (*registerpb.RegRes, error) {
	return logic.NewRegisterByPhoneLogic(ctx, s.svcCtx).RegisterByPhone(in)
}

// SendCode 发送注册验证码
// 接收 gRPC 请求并转发给 SendCodeLogic 处理
//
// 参数：
//   - ctx: 请求上下文
//   - in: 验证码请求，包含手机号
//
// 返回：
//   - NoRes: 空响应
//   - error: 错误信息
func (s *RegisterServer) SendCode(ctx context.Context, in *registerpb.CodeReq) (*registerpb.NoRes, error) {
	return logic.NewSendCodeLogic(ctx, s.svcCtx).SendCode(in)
}
