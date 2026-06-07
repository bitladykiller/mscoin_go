package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

type SendWithdrawCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendWithdrawCodeLogic {
	return &SendWithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *SendWithdrawCodeLogic) SendCode(req *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	if err := l.svcCtx.WithdrawService.SendCode(l.ctx, req.Phone); err != nil {
		return nil, err
	}
	return &withdrawpb.NoRes{}, nil
}
