package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

// FindMemberByIDLogic 根据 ID 查询会员信息的逻辑处理器
type FindMemberByIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFindMemberByIDLogic 创建逻辑处理器实例
func NewFindMemberByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindMemberByIDLogic {
	return &FindMemberByIDLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindMemberByID 根据会员 ID 查询会员信息
func (l *FindMemberByIDLogic) FindMemberByID(req *memberpb.MemberReq) (*memberpb.MemberInfo, error) {
	return l.svcCtx.MemberService.FindByID(l.ctx, req.MemberId)
}
