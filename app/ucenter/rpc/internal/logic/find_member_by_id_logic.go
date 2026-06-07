package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

type FindMemberByIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindMemberByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindMemberByIDLogic {
	return &FindMemberByIDLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *FindMemberByIDLogic) FindMemberByID(req *memberpb.MemberReq) (*memberpb.MemberInfo, error) {
	return l.svcCtx.MemberService.FindByID(l.ctx, req.MemberId)
}
