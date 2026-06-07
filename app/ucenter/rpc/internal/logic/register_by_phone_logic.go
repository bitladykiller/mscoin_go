package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// RegisterByPhoneLogic 手机号注册逻辑处理器
type RegisterByPhoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewRegisterByPhoneLogic 创建逻辑处理器实例
func NewRegisterByPhoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterByPhoneLogic {
	return &RegisterByPhoneLogic{ctx: ctx, svcCtx: svcCtx}
}

// RegisterByPhone 处理手机号注册请求
func (l *RegisterByPhoneLogic) RegisterByPhone(req *registerpb.RegReq) (*registerpb.RegRes, error) {
	return l.svcCtx.MemberService.RegisterByPhone(l.ctx, req)
}
