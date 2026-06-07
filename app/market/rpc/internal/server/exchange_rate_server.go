package server

import (
	"context"

	"mscoin_go/app/market/rpc/internal/logic"
	"mscoin_go/app/market/rpc/internal/svc"
	ratepb "mscoin_go/app/market/rpc/pb/rate"
)

// ExchangeRateServer 是法币汇率查询的 RPC 门面。
type ExchangeRateServer struct {
	svcCtx *svc.ServiceContext
	ratepb.UnimplementedExchangeRateServer
}

func NewExchangeRateServer(svcCtx *svc.ServiceContext) *ExchangeRateServer {
	return &ExchangeRateServer{svcCtx: svcCtx}
}

func (s *ExchangeRateServer) UsdRate(ctx context.Context, req *ratepb.RateReq) (*ratepb.RateRes, error) {
	return logic.NewUsdRateLogic(ctx, s.svcCtx).UsdRate(req)
}
