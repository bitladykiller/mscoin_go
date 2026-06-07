package logic

import (
	"context"
	"errors"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type CoinInfoLogic struct {
	marketLogicBase
}

func NewCoinInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CoinInfoLogic {
	return &CoinInfoLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *CoinInfoLogic) CoinInfo(req *types.MarketReq) (*types.Coin, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	coin, err := l.svcCtx.MarketClient.FindCoinInfo(ctx, &marketpb.MarketReq{Unit: req.Unit})
	if err != nil {
		return nil, err
	}

	resp := &types.Coin{}
	if err := copier.Copy(resp, coin); err != nil {
		return nil, errors.New("market coin payload copy failed")
	}
	return resp, nil
}
