package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

type WithdrawCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawCodeLogic {
	return &WithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *WithdrawCodeLogic) WithdrawCode(req *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	if err := l.svcCtx.WithdrawService.Apply(l.ctx, req); err != nil {
		return nil, err
	}
	return &withdrawpb.NoRes{}, nil
}
