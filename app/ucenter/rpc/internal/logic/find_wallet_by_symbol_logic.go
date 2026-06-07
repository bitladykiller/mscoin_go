package logic

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindWalletBySymbolLogic 根据币种查询钱包的逻辑处理器
type FindWalletBySymbolLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFindWalletBySymbolLogic 创建逻辑处理器实例
func NewFindWalletBySymbolLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindWalletBySymbolLogic {
	return &FindWalletBySymbolLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindWalletBySymbol 根据币种查询会员钱包
func (l *FindWalletBySymbolLogic) FindWalletBySymbol(req *assetpb.AssetReq) (*assetpb.MemberWallet, error) {
	coin, err := l.svcCtx.MarketClient.FindCoinInfo(l.ctx, &marketpb.MarketReq{Unit: req.CoinName})
	if err != nil {
		return nil, err
	}
	return l.svcCtx.WalletService.FindWalletBySymbol(l.ctx, req.UserId, req.CoinName, coin)
}
