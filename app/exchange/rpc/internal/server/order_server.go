package server

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/logic"
	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

type OrderServer struct {
	svcCtx *svc.ServiceContext
	orderpb.UnimplementedOrderServer
}

func NewOrderServer(svcCtx *svc.ServiceContext) *OrderServer {
	return &OrderServer{svcCtx: svcCtx}
}

func (s *OrderServer) FindOrderHistory(ctx context.Context, req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	return logic.NewFindOrderHistoryLogic(ctx, s.svcCtx).FindOrderHistory(req)
}

func (s *OrderServer) FindOrderCurrent(ctx context.Context, req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	return logic.NewFindOrderCurrentLogic(ctx, s.svcCtx).FindOrderCurrent(req)
}

func (s *OrderServer) Add(ctx context.Context, req *orderpb.OrderReq) (*orderpb.AddOrderRes, error) {
	return logic.NewAddLogic(ctx, s.svcCtx).Add(req)
}

func (s *OrderServer) FindByOrderId(ctx context.Context, req *orderpb.OrderReq) (*orderpb.ExchangeOrderOrigin, error) {
	return logic.NewFindByOrderIDLogic(ctx, s.svcCtx).FindByOrderID(req)
}

func (s *OrderServer) CancelOrder(ctx context.Context, req *orderpb.OrderReq) (*orderpb.CancelOrderRes, error) {
	return logic.NewCancelOrderLogic(ctx, s.svcCtx).CancelOrder(req)
}
