package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

type RegisterServer struct {
	svcCtx *svc.ServiceContext
	registerpb.UnimplementedRegisterServer
}

func NewRegisterServer(svcCtx *svc.ServiceContext) *RegisterServer {
	return &RegisterServer{svcCtx: svcCtx}
}

func (s *RegisterServer) RegisterByPhone(ctx context.Context, in *registerpb.RegReq) (*registerpb.RegRes, error) {
	return logic.NewRegisterByPhoneLogic(ctx, s.svcCtx).RegisterByPhone(in)
}

func (s *RegisterServer) SendCode(ctx context.Context, in *registerpb.CodeReq) (*registerpb.NoRes, error) {
	return logic.NewSendCodeLogic(ctx, s.svcCtx).SendCode(in)
}
