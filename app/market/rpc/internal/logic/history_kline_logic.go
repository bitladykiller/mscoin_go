package logic

import (
	"context"
	"time"

	"mscoin_go/app/market/rpc/internal/svc"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type HistoryKlineLogic struct {
	marketLogicBase
}

func NewHistoryKlineLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryKlineLogic {
	return &HistoryKlineLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *HistoryKlineLogic) HistoryKline(req *marketpb.MarketReq) (*marketpb.HistoryRes, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	list, err := l.marketService.HistoryKline(ctx, req.Symbol, req.From, req.To, req.Resolution)
	if err != nil {
		return nil, err
	}

	return &marketpb.HistoryRes{List: list}, nil
}
