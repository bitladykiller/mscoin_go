package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type FindCoinByIDLogic struct {
	marketLogicBase
}

func NewFindCoinByIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindCoinByIDLogic {
	return &FindCoinByIDLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *FindCoinByIDLogic) FindByID(req *marketpb.MarketReq) (*marketpb.Coin, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	coin, err := l.coinService.FindCoinByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &marketpb.Coin{}
	if err := copier.Copy(resp, coin); err != nil {
		return nil, err
	}
	return resp, nil
}
