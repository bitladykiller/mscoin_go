package logic

import (
	"context"
	"errors"

	"mscoin_go/app/exchange/api/internal/middleware"
	"mscoin_go/app/exchange/api/internal/svc"
	"mscoin_go/app/exchange/api/internal/types"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

type AddLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *AddLogic) Add(req *types.ExchangeReq) (string, error) {
	if !req.OrderValid() {
		return "", errors.New("invalid request")
	}

	userID := middleware.UserIDFromContext(l.ctx)
	resp, err := l.svcCtx.OrderClient.Add(l.ctx, &orderpb.OrderReq{
		Symbol:    req.Symbol,
		UserId:    userID,
		Direction: req.Direction,
		Type:      req.Type,
		Price:     req.Price,
		Amount:    req.Amount,
	})
	if err != nil {
		return "", err
	}
	return resp.OrderId, nil
}
