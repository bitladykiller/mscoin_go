package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// FindByOrderIDLogic 处理根据订单 ID 查询订单的 RPC 请求。
type FindByOrderIDLogic struct {
	exchangeLogicBase
}

// NewFindByOrderIDLogic 创建 FindByOrderIDLogic 实例。
func NewFindByOrderIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindByOrderIDLogic {
	return &FindByOrderIDLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// FindByOrderID 执行根据订单 ID 查询订单的业务逻辑。
// 返回订单的完整信息（原始格式，包含数值型的状态、方向、类型）。
func (l *FindByOrderIDLogic) FindByOrderID(req *orderpb.OrderReq) (*orderpb.ExchangeOrderOrigin, error) {
	// 调用服务层查询订单
	order, err := l.svcCtx.OrderService.FindByOrderID(l.ctx, req.OrderId)
	if err != nil {
		return nil, err
	}

	// 将订单实体转换为 RPC 响应格式
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
