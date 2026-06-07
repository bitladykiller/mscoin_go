package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

type FindOrderCurrentLogic struct {
	exchangeLogicBase
}

func NewFindOrderCurrentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindOrderCurrentLogic {
	return &FindOrderCurrentLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

func (l *FindOrderCurrentLogic) FindOrderCurrent(req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	list, total, err := l.svcCtx.OrderService.FindOrderCurrent(l.ctx, req.Symbol, req.Page, req.PageSize, req.UserId)
	if err != nil {
		return nil, err
	}

	resp := make([]*orderpb.ExchangeOrder, len(list))
	if err := copier.Copy(&resp, list); err != nil {
		return nil, err
	}
	return &orderpb.OrderRes{List: resp, Total: total}, nil
}
