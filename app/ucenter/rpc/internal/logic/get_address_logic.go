package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type GetAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAddressLogic {
	return &GetAddressLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *GetAddressLogic) GetAddress(req *assetpb.AssetReq) (*assetpb.AddressList, error) {
	addresses, err := l.svcCtx.WalletService.GetAllAddress(l.ctx, req.CoinName)
	if err != nil {
		return nil, err
	}
	return &assetpb.AddressList{List: addresses}, nil
}
