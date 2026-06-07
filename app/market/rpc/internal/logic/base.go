package logic

import (
	"context"

	"mscoin_go/app/market/rpc/internal/domain/service"
	"mscoin_go/app/market/rpc/internal/svc"
)

// marketLogicBase collects the domain services used by the market RPC logic
// files. This keeps each generated-style logic file small while still making
// dependencies explicit.
type marketLogicBase struct {
	ctx                 context.Context
	svcCtx              *svc.ServiceContext
	coinService         *service.CoinService
	exchangeCoinService *service.ExchangeCoinService
	marketService       *service.MarketService
	rateService         *service.RateService
}

func newMarketLogicBase(ctx context.Context, svcCtx *svc.ServiceContext) marketLogicBase {
	return marketLogicBase{
		ctx:                 ctx,
		svcCtx:              svcCtx,
		coinService:         svcCtx.CoinService,
		exchangeCoinService: svcCtx.ExchangeCoinService,
		marketService:       svcCtx.MarketService,
		rateService:         svcCtx.RateService,
	}
}
