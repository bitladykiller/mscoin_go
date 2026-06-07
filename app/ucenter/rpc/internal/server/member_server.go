package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

// MemberServer 会员 RPC 服务端
// 处理会员信息相关的 RPC 请求
type MemberServer struct {
	svcCtx *svc.ServiceContext
	memberpb.UnimplementedMemberServer
}

// NewMemberServer 创建会员服务端实例
func NewMemberServer(svcCtx *svc.ServiceContext) *MemberServer {
	return &MemberServer{svcCtx: svcCtx}
}

// FindMemberById 根据会员 ID 查询会员信息
func (s *MemberServer) FindMemberById(ctx context.Context, in *memberpb.MemberReq) (*memberpb.MemberInfo, error) {
	return logic.NewFindMemberByIDLogic(ctx, s.svcCtx).FindMemberByID(in)
}
