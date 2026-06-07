package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

type FindByOrderIDLogic struct {
	exchangeLogicBase
}

func NewFindByOrderIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindByOrderIDLogic {
	return &FindByOrderIDLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

func (l *FindByOrderIDLogic) FindByOrderID(req *orderpb.OrderReq) (*orderpb.ExchangeOrderOrigin, error) {
	order, err := l.svcCtx.OrderService.FindByOrderID(l.ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	return &orderpb.ExchangeOrderOrigin{
		Id:            order.ID,
		OrderId:       order.OrderId,
		Amount:        order.Amount,
		BaseSymbol:    order.BaseSymbol,
		CanceledTime:  order.CanceledTime,
		CoinSymbol:    order.CoinSymbol,
		CompletedTime: order.CompletedTime,
		Direction:     int32(order.Direction),
		MemberId:      order.MemberId,
		Price:         order.Price,
		Status:        int32(order.Status),
		Symbol:        order.Symbol,
		Time:          order.Time,
		TradedAmount:  order.TradedAmount,
		Turnover:      order.Turnover,
		Type:          int32(order.Type),
		UseDiscount:   order.UseDiscount,
	}, nil
}
