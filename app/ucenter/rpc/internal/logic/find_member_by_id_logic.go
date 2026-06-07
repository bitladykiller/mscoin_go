// Package logic 定义会员查询业务逻辑处理器。
package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	memberpb "mscoin_go/app/ucenter/rpc/pb/member"
)

// FindMemberByIDLogic 根据 ID 查询会员信息的逻辑处理器
// 处理会员信息查询 RPC 请求
type FindMemberByIDLogic struct {
	ctx    context.Context     // 请求上下文
	svcCtx *svc.ServiceContext // 服务上下文
}

// NewFindMemberByIDLogic 创建逻辑处理器实例
func NewFindMemberByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindMemberByIDLogic {
	return &FindMemberByIDLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindMemberByID 根据会员 ID 查询会员信息
// 调用 MemberService.FindByID 处理查询逻辑
//
// 参数：
//   - req: 会员请求，包含会员 ID
//
// 返回：
//   - MemberInfo: 会员信息（完整字段）
//   - error: 错误信息
func (l *FindMemberByIDLogic) FindMemberByID(req *memberpb.MemberReq) (*memberpb.MemberInfo, error) {
	return l.svcCtx.MemberService.FindByID(l.ctx, req.MemberId)
}
