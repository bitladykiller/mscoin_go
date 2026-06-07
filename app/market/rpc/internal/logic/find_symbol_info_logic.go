package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type FindSymbolInfoLogic struct {
	marketLogicBase
}

func NewFindSymbolInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindSymbolInfoLogic {
	return &FindSymbolInfoLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *FindSymbolInfoLogic) FindSymbolInfo(req *marketpb.MarketReq) (*marketpb.ExchangeCoin, error) {
	item, err := l.exchangeCoinService.FindBySymbol(l.ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	resp := &marketpb.ExchangeCoin{}
	if err := copier.Copy(resp, item); err != nil {
		return nil, err
	}
	return resp, nil
}
