package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

type FindAddressByCoinIDLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindAddressByCoinIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindAddressByCoinIDLogic {
	return &FindAddressByCoinIDLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *FindAddressByCoinIDLogic) FindAddressByCoinID(req *withdrawpb.WithdrawReq) (*withdrawpb.AddressSimpleList, error) {
	list, err := l.svcCtx.WithdrawService.FindAddressByCoinID(l.ctx, req.UserId, req.CoinId)
	if err != nil {
		return nil, err
	}
	return &withdrawpb.AddressSimpleList{List: list}, nil
}
