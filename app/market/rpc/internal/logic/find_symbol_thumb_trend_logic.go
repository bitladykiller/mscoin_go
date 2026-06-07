package logic

import (
	"context"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type FindSymbolThumbTrendLogic struct {
	marketLogicBase
}

func NewFindSymbolThumbTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindSymbolThumbTrendLogic {
	return &FindSymbolThumbTrendLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *FindSymbolThumbTrendLogic) FindSymbolThumbTrend(*marketpb.MarketReq) (*marketpb.SymbolThumbRes, error) {
	list, err := l.marketService.SymbolThumbTrend(l.ctx)
	if err != nil {
		return nil, err
	}

	return &marketpb.SymbolThumbRes{List: list}, nil
}
