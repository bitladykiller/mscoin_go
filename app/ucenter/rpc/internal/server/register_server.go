package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// RegisterServer 注册 RPC 服务端
// 处理会员注册相关的 RPC 请求
type RegisterServer struct {
	svcCtx *svc.ServiceContext
	registerpb.UnimplementedRegisterServer
}

// NewRegisterServer 创建注册服务端实例
func NewRegisterServer(svcCtx *svc.ServiceContext) *RegisterServer {
	return &RegisterServer{svcCtx: svcCtx}
}

// RegisterByPhone 处理手机号注册请求
func (s *RegisterServer) RegisterByPhone(ctx context.Context, in *registerpb.RegReq) (*registerpb.RegRes, error) {
	return logic.NewRegisterByPhoneLogic(ctx, s.svcCtx).RegisterByPhone(in)
}

// SendCode 发送注册验证码
func (s *RegisterServer) SendCode(ctx context.Context, in *registerpb.CodeReq) (*registerpb.NoRes, error) {
	return logic.NewSendCodeLogic(ctx, s.svcCtx).SendCode(in)
}
