package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
)

type LoginServer struct {
	svcCtx *svc.ServiceContext
	loginpb.UnimplementedLoginServer
}

func NewLoginServer(svcCtx *svc.ServiceContext) *LoginServer {
	return &LoginServer{svcCtx: svcCtx}
}

func (s *LoginServer) Login(ctx context.Context, in *loginpb.LoginReq) (*loginpb.LoginRes, error) {
	return logic.NewLoginLogic(ctx, s.svcCtx).Login(in)
}
