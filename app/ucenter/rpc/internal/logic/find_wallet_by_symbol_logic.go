package logic

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type FindWalletBySymbolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindWalletBySymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletBySymbolLogic {
	return &FindWalletBySymbolLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *FindWalletBySymbolLogic) FindWalletBySymbol(req *assetpb.AssetReq) (*assetpb.MemberWallet, error) {
	coin, err := l.svcCtx.MarketClient.FindCoinInfo(l.ctx, &marketpb.MarketReq{Unit: req.CoinName})
	if err != nil {
		return nil, err
	}
	return l.svcCtx.WalletService.FindWalletBySymbol(l.ctx, req.UserId, req.CoinName, coin)
}
