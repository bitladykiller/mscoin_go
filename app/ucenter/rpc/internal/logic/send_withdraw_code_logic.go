package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// SendWithdrawCodeLogic 发送提现验证码逻辑处理器
type SendWithdrawCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendWithdrawCodeLogic 创建逻辑处理器实例
func NewSendWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawCodeLogic {
	return &SendWithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// SendCode 发送提现验证码
func (l *SendWithdrawCodeLogic) SendCode(req *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	if err := l.svcCtx.WithdrawService.SendCode(l.ctx, req.Phone); err != nil {
		return nil, err
	}
	return &withdrawpb.NoRes{}, nil
}
