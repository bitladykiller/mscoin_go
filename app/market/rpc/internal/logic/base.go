package logic

import (
	"context"

	"mscoin_go/app/market/rpc/internal/domain/service"
	"mscoin_go/app/market/rpc/internal/svc"
)

// marketLogicBase 聚合 market RPC logic 文件中使用的领域服务。
// 这样既保持每个生成式 logic 文件精简，又使依赖关系显式化。
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
