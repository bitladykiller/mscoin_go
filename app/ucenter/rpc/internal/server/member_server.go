// Package server 定义会员 RPC 服务端。
package server

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/logic"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

// MemberServer 会员 RPC 服务端
// 处理会员信息相关的 RPC 请求
//
// 实现 memberpb.MemberServer 接口
// 提供 FindMemberById 方法处理会员信息查询
type MemberServer struct {
	svcCtx *svc.ServiceContext         // 服务上下文
	memberpb.UnimplementedMemberServer // 未实现方法的默认实现
}

// NewMemberServer 创建会员服务端实例
func NewMemberServer(svcCtx *svc.ServiceContext) *MemberServer {
	return &MemberServer{svcCtx: svcCtx}
}

// FindMemberById 根据会员 ID 查询会员信息
// 接收 gRPC 请求并转发给 FindMemberByIDLogic 处理
//
// 参数：
//   - ctx: 请求上下文
//   - in: 会员请求，包含会员 ID
//
// 返回：
//   - MemberInfo: 会员信息（完整字段）
//   - error: 错误信息
func (s *MemberServer) FindMemberById(ctx context.Context, in *memberpb.MemberReq) (*memberpb.MemberInfo, error) {
	return logic.NewFindMemberByIDLogic(ctx, s.svcCtx).FindMemberByID(in)
}
