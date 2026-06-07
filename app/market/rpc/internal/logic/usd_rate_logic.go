package logic

import (
	"context"

	"mscoin_go/app/market/rpc/internal/svc"
	ratepb "mscoin_go/app/market/rpc/pb/rate"
)

type UsdRateLogic struct {
	marketLogicBase
}

func NewUsdRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UsdRateLogic {
	return &UsdRateLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *UsdRateLogic) UsdRate(req *ratepb.RateReq) (*ratepb.RateRes, error) {
	return &ratepb.RateRes{Rate: l.rateService.USDRate(l.ctx, req.Unit)}, nil
}
