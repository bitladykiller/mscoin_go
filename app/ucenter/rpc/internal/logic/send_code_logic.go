package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	registerpb "mscoin_go/app/ucenter/rpc/pb/register"
)

// SendCodeLogic 发送验证码逻辑处理器
type SendCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendCodeLogic 创建逻辑处理器实例
func NewSendCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCodeLogic {
	return &SendCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendCode 发送注册验证码
func (l *SendCodeLogic) SendCode(req *registerpb.CodeReq) (*registerpb.NoRes, error) {
	return l.svcCtx.MemberService.SendRegisterCode(l.ctx, req)
}
