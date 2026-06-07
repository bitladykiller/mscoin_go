package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
)

// LoginServer 登录 RPC 服务端
// 处理会员登录相关的 RPC 请求
type LoginServer struct {
	svcCtx *svc.ServiceContext
	loginpb.UnimplementedLoginServer
}

// NewLoginServer 创建登录服务端实例
func NewLoginServer(svcCtx *svc.ServiceContext) *LoginServer {
	return &LoginServer{svcCtx: svcCtx}
}

// Login 处理会员登录请求
func (s *LoginServer) Login(ctx context.Context, in *loginpb.LoginReq) (*loginpb.LoginRes, error) {
	return logic.NewLoginLogic(ctx, s.svcCtx).Login(in)
}
