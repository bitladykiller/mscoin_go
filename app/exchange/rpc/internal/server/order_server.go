// Package server 定义了 exchange-rpc 的 gRPC 服务器实现。
// OrderServer 实现了 Order 服务接口，将 RPC 请求路由到对应的业务逻辑处理器。
//
// gRPC 服务方法：
// - FindOrderHistory: 查询历史订单
// - FindOrderCurrent: 查询当前委托订单
// - Add: 新增订单
// - FindByOrderId: 根据 ID 查询订单
// - CancelOrder: 取消订单
//
// 架构说明：
// - OrderServer 作为 gRPC 服务的入口点
// - 每个方法创建对应的 Logic 实例处理业务逻辑
// - Logic 实例通过 ServiceContext 访问数据库和外部 RPC 服务
package server

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/logic"
	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// OrderServer 是 Order 服务的 gRPC 服务器实现。
// 实现了 orderpb.OrderServer 接口定义的所有方法。
//
// 服务方法与 Logic 的对应关系：
// - FindOrderHistory -> FindOrderHistoryLogic
// - FindOrderCurrent -> FindOrderCurrentLogic
// - Add -> AddLogic
// - FindByOrderId -> FindByOrderIDLogic
// - CancelOrder -> CancelOrderLogic
type OrderServer struct {
	svcCtx *svc.ServiceContext
	orderpb.UnimplementedOrderServer
}

// NewOrderServer 创建 OrderServer 实例。
func NewOrderServer(svcCtx *svc.ServiceContext) *OrderServer {
	return &OrderServer{svcCtx: svcCtx}
}

// FindOrderHistory 查询历史订单 RPC 方法。
// 返回用户已完成或已取消的历史订单列表。
//
// 调用流程：
// - 创建 FindOrderHistoryLogic 实例
// - 调用 FindOrderHistory 方法查询数据库
// - 返回分页结果
func (s *OrderServer) FindOrderHistory(ctx context.Context, req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	return logic.NewFindOrderHistoryLogic(ctx, s.svcCtx).FindOrderHistory(req)
}

// FindOrderCurrent 查询当前订单 RPC 方法。
// 返回用户当前正在委托中的订单列表（TRADING 状态）。
//
// 调用流程：
// - 创建 FindOrderCurrentLogic 实例
// - 调用 FindOrderCurrent 方法查询数据库
// - 返回分页结果
func (s *OrderServer) FindOrderCurrent(ctx context.Context, req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	return logic.NewFindOrderCurrentLogic(ctx, s.svcCtx).FindOrderCurrent(req)
}

// Add 新增订单 RPC 方法。
// 创建新订单，执行完整的业务规则验证。
//
// 调用流程：
// - 创建 AddLogic 实例
// - 调用 Add 方法创建订单
//   - 调用 ucenter-rpc 查询会员信息和钱包信息
//   - 调用 market-rpc 查询交易对配置
//   - 调用 OrderService 验证业务规则并保存订单
// - 返回订单 ID
func (s *OrderServer) Add(ctx context.Context, req *orderpb.OrderReq) (*orderpb.AddOrderRes, error) {
	return logic.NewAddLogic(ctx, s.svcCtx).Add(req)
}

// FindByOrderId 根据订单 ID 查询订单 RPC 方法。
// 返回订单的原始信息，包含数值型的状态、方向、类型字段。
//
// 调用流程：
// - 创建 FindByOrderIDLogic 实例
// - 调用 FindByOrderID 方法查询数据库
// - 返回订单原始信息
func (s *OrderServer) FindByOrderId(ctx context.Context, req *orderpb.OrderReq) (*orderpb.ExchangeOrderOrigin, error) {
	return logic.NewFindByOrderIDLogic(ctx, s.svcCtx).FindByOrderID(req)
}

// CancelOrder 取消订单 RPC 方法。
// 将订单状态更新为 CANCELED。
//
// 调用流程：
// - 创建 CancelOrderLogic 实例
// - 调用 CancelOrder 方法更新订单状态
// - 返回取消结果
//
// 注意：取消订单后，实际应用中应释放冻结的资金（调用 ucenter-rpc）。
func (s *OrderServer) CancelOrder(ctx context.Context, req *orderpb.OrderReq) (*orderpb.CancelOrderRes, error) {
	return logic.NewCancelOrderLogic(ctx, s.svcCtx).CancelOrder(req)
}
