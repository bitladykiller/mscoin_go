package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type FindExchangeCoinVisibleLogic struct {
	marketLogicBase
}

func NewFindExchangeCoinVisibleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindExchangeCoinVisibleLogic {
	return &FindExchangeCoinVisibleLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *FindExchangeCoinVisibleLogic) FindExchangeCoinVisible(*marketpb.MarketReq) (*marketpb.ExchangeCoinRes, error) {
	list, err := l.exchangeCoinService.FindVisible(l.ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]*marketpb.ExchangeCoin, len(list))
	if err := copier.Copy(&resp, list); err != nil {
		return nil, err
	}

	return &marketpb.ExchangeCoinRes{List: resp}, nil
}
