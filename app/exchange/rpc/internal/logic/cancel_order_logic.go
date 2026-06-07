package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// CancelOrderLogic 处理取消订单的 RPC 请求。
// 将订单状态更新为已取消。
type CancelOrderLogic struct {
	exchangeLogicBase
}

// NewCancelOrderLogic 创建 CancelOrderLogic 实例。
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// CancelOrder 执行取消订单的业务逻辑。
//
// 处理流程：
// 1. 调用 OrderService.CancelOrder 更新订单状态
// 2. 返回取消结果（订单 ID）
//
// 取消订单说明：
// - 将订单状态从 TRADING 更新为 CANCELED
// - 取消后订单不再参与撮合
// - 取消后的订单会出现在历史订单列表中
//
// 业务规则验证（应在调用方完成）：
// - 验证订单是否存在
// - 验证订单是否属于当前用户
// - 验证订单状态是否为 TRADING（只有交易中的订单才能取消）
//
// 注意：本方法不调用 ucenter-rpc 或 market-rpc，直接操作本地数据库。
// 实际应用中，取消订单后应调用 ucenter-rpc 释放冻结的资金。
func (l *CancelOrderLogic) CancelOrder(req *orderpb.OrderReq) (*orderpb.CancelOrderRes, error) {
	// 调用服务层取消订单
	if err := l.svcCtx.OrderService.CancelOrder(l.ctx, req.OrderId); err != nil {
		return nil, err
	}
	return &orderpb.CancelOrderRes{OrderId: req.OrderId}, nil
}
