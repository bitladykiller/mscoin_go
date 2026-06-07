package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// FindByOrderIDLogic 处理根据订单 ID 查询订单的 RPC 请求。
// 返回订单的完整原始信息，包含数值型的状态、方向、类型字段。
type FindByOrderIDLogic struct {
	exchangeLogicBase
}

// NewFindByOrderIDLogic 创建 FindByOrderIDLogic 实例。
func NewFindByOrderIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindByOrderIDLogic {
	return &FindByOrderIDLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// FindByOrderID 执行根据订单 ID 查询订单的业务逻辑。
//
// 处理流程：
// 1. 调用 OrderService.FindByOrderID 查询数据库
// 2. 将订单实体转换为 ExchangeOrderOrigin 格式
// 3. 返回订单原始信息
//
// 返回数据说明：
// - ExchangeOrderOrigin 包含数值型的状态、方向、类型字段
// - 状态：0-交易中, 1-已完成, 2-已取消, 3-已超时, 4-初始化
// - 方向：0-买入, 1-卖出
// - 类型：0-市价, 1-限价
//
// 与其他查询方法的区别：
// - FindOrderHistory/FindOrderCurrent 返回 ExchangeOrder，状态/方向/类型为字符串标签
// - FindByOrderID 返回 ExchangeOrderOrigin，状态/方向/类型为数值代码
// - 数值代码便于内部处理，字符串标签便于前端展示
//
// 注意：本方法不调用 ucenter-rpc 或 market-rpc，直接查询本地数据库。
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
