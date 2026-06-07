package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// CancelOrderLogic 处理取消订单的 RPC 请求。
type CancelOrderLogic struct {
	exchangeLogicBase
}

// NewCancelOrderLogic 创建 CancelOrderLogic 实例。
func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// CancelOrder 执行取消订单的业务逻辑。
// 将订单状态更新为已取消。
func (l *CancelOrderLogic) CancelOrder(req *orderpb.OrderReq) (*orderpb.CancelOrderRes, error) {
	// 调用服务层取消订单
	if err := l.svcCtx.OrderService.CancelOrder(l.ctx, req.OrderId); err != nil {
		return nil, err
	}
	return &orderpb.CancelOrderRes{OrderId: req.OrderId}, nil
}
