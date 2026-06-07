package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

type MemberServer struct {
	svcCtx *svc.ServiceContext
	memberpb.UnimplementedMemberServer
}

func NewMemberServer(svcCtx *svc.ServiceContext) *MemberServer {
	return &MemberServer{svcCtx: svcCtx}
}

func (s *MemberServer) FindMemberById(ctx context.Context, in *memberpb.MemberReq) (*memberpb.MemberInfo, error) {
	return logic.NewFindMemberByIDLogic(ctx, s.svcCtx).FindMemberByID(in)
}
