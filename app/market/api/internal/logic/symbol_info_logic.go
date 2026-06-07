package logic

import (
	"context"
	"time"

	"github.com/jinzhu/copier"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	marketpb "mscoin_go/app/market/rpc/pb/market"
)

type SymbolInfoLogic struct {
	marketLogicBase
}

func NewSymbolInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SymbolInfoLogic {
	return &SymbolInfoLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *SymbolInfoLogic) SymbolInfo(req types.MarketReq) (*types.ExchangeCoinResp, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	payload, err := l.svcCtx.MarketClient.FindSymbolInfo(ctx, &marketpb.MarketReq{
		Ip:     req.IP,
		Symbol: req.Symbol,
	})
	if err != nil {
		return nil, err
	}

	resp := &types.ExchangeCoinResp{}
	if err := copier.Copy(resp, payload); err != nil {
		return nil, err
	}
	return resp, nil
}
