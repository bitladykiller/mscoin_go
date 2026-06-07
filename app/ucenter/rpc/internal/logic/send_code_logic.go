package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

type SendCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *SendCodeLogic) SendCode(req *registerpb.CodeReq) (*registerpb.NoRes, error) {
	return l.svcCtx.MemberService.SendRegisterCode(l.ctx, req)
}
