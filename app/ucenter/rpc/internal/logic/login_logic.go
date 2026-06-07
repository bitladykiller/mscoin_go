package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	loginpb "mscoin_go/app/ucenter/rpc/pb/login"
)

// LoginLogic 登录逻辑处理器
type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewLoginLogic 创建逻辑处理器实例
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{ctx: ctx, svcCtx: svcCtx}
}

// Login 处理会员登录请求
func (l *LoginLogic) Login(req *loginpb.LoginReq) (*loginpb.LoginRes, error) {
	return l.svcCtx.MemberService.Login(l.ctx, req)
}
