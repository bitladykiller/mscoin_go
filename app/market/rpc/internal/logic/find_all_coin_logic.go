package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type FindAllCoinLogic struct {
	marketLogicBase
}

func NewFindAllCoinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindAllCoinLogic {
	return &FindAllCoinLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *FindAllCoinLogic) FindAllCoin(*marketpb.MarketReq) (*marketpb.CoinList, error) {
	list, err := l.coinService.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]*marketpb.Coin, len(list))
	if err := copier.Copy(&resp, list); err != nil {
		return nil, err
	}

	return &marketpb.CoinList{List: resp}, nil
}
