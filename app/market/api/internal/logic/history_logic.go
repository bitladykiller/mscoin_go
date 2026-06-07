package logic

import (
	"context"
	"time"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type HistoryLogic struct {
	marketLogicBase
}

func NewHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryLogic {
	return &HistoryLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *HistoryLogic) History(req *types.MarketReq) (*types.HistoryKline, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	payload, err := l.svcCtx.MarketClient.HistoryKline(ctx, &marketpb.MarketReq{
		Symbol:     req.Symbol,
		From:       req.From,
		To:         req.To,
		Resolution: req.Resolution,
	})
	if err != nil {
		return nil, err
	}

	list := make([][]any, len(payload.List))
	for i, item := range payload.List {
		list[i] = []any{item.Time, item.Open, item.High, item.Low, item.Close, item.Volume}
	}

	return &types.HistoryKline{List: list}, nil
}
