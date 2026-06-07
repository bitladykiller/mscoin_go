package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

type RegisterByPhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterByPhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterByPhoneLogic {
	return &RegisterByPhoneLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *RegisterByPhoneLogic) RegisterByPhone(req *registerpb.RegReq) (*registerpb.RegRes, error) {
	return l.svcCtx.MemberService.RegisterByPhone(l.ctx, req)
}
