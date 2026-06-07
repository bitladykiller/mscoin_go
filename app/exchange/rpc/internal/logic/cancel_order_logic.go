package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

type CancelOrderLogic struct {
	exchangeLogicBase
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

func (l *CancelOrderLogic) CancelOrder(req *orderpb.OrderReq) (*orderpb.CancelOrderRes, error) {
	if err := l.svcCtx.OrderService.CancelOrder(l.ctx, req.OrderId); err != nil {
		return nil, err
	}
	return &orderpb.CancelOrderRes{OrderId: req.OrderId}, nil
}
