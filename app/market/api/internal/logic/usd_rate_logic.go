package logic

import (
	"context"
	"time"

	"mscoin_go/app/market/api/internal/svc"
	"mscoin_go/app/market/api/internal/types"
	ratepb "mscoin_go/app/market/rpc/pb/rate"
)

type UsdRateLogic struct {
	marketLogicBase
}

func NewUsdRateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UsdRateLogic {
	return &UsdRateLogic{marketLogicBase: newMarketLogicBase(ctx, svcCtx)}
}

func (l *UsdRateLogic) UsdRate(req *types.RateRequest) (*types.RateResponse, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	payload, err := l.svcCtx.RateClient.UsdRate(ctx, &ratepb.RateReq{
		Unit: req.Unit,
		Ip:   req.IP,
	})
	if err != nil {
		return nil, err
	}

	return &types.RateResponse{Rate: payload.Rate}, nil
}
