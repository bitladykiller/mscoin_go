// Package server 定义了 exchange-rpc 的 gRPC 服务器实现。
// OrderServer 实现了 Order 服务接口，将 RPC 请求路由到对应的业务逻辑处理器。
package server

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/logic"
	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// OrderServer 是 Order 服务的 gRPC 服务器实现。
type OrderServer struct {
	svcCtx *svc.ServiceContext
	orderpb.UnimplementedOrderServer
}

// NewOrderServer 创建 OrderServer 实例。
func NewOrderServer(svcCtx *svc.ServiceContext) *OrderServer {
	return &OrderServer{svcCtx: svcCtx}
}

// FindOrderHistory 查询历史订单 RPC 方法。
func (s *OrderServer) FindOrderHistory(ctx context.Context, req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	return logic.NewFindOrderHistoryLogic(ctx, s.svcCtx).FindOrderHistory(req)
}

// FindOrderCurrent 查询当前订单 RPC 方法。
func (s *OrderServer) FindOrderCurrent(ctx context.Context, req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	return logic.NewFindOrderCurrentLogic(ctx, s.svcCtx).FindOrderCurrent(req)
}

// Add 新增订单 RPC 方法。
func (s *OrderServer) Add(ctx context.Context, req *orderpb.OrderReq) (*orderpb.AddOrderRes, error) {
	return logic.NewAddLogic(ctx, s.svcCtx).Add(req)
}

// FindByOrderId 根据订单 ID 查询订单 RPC 方法。
func (s *OrderServer) FindByOrderId(ctx context.Context, req *orderpb.OrderReq) (*orderpb.ExchangeOrderOrigin, error) {
	return logic.NewFindByOrderIDLogic(ctx, s.svcCtx).FindByOrderID(req)
}

// CancelOrder 取消订单 RPC 方法。
func (s *OrderServer) CancelOrder(ctx context.Context, req *orderpb.OrderReq) (*orderpb.CancelOrderRes, error) {
	return logic.NewCancelOrderLogic(ctx, s.svcCtx).CancelOrder(req)
}
