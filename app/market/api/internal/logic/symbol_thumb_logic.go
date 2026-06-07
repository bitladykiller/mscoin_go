package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type SymbolThumbLogic struct {
	marketLogicBase
}

func NewSymbolThumbLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolThumbLogic {
	return &SymbolThumbLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *SymbolThumbLogic) SymbolThumb(req *types.MarketReq) ([]*types.CoinThumbResp, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	payload, err := l.svcCtx.MarketClient.FindSymbolThumbTrend(ctx, &marketpb.MarketReq{Ip: req.IP})
	if err != nil {
		return nil, err
	}

	list := make([]*types.CoinThumbResp, len(payload.List))
	if err := copier.Copy(&list, payload.List); err != nil {
		return nil, err
	}
	return list, nil
}
