package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// FindOrderCurrentLogic 处理查询当前订单的 RPC 请求。
type FindOrderCurrentLogic struct {
	exchangeLogicBase
}

// NewFindOrderCurrentLogic 创建 FindOrderCurrentLogic 实例。
func NewFindOrderCurrentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindOrderCurrentLogic {
	return &FindOrderCurrentLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// FindOrderCurrent 执行查询当前订单的业务逻辑。
// 返回用户当前委托中的订单列表（交易中状态）。
func (l *FindOrderCurrentLogic) FindOrderCurrent(req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	// 调用服务层查询当前订单
	list, total, err := l.svcCtx.OrderService.FindOrderCurrent(l.ctx, req.Symbol, req.Page, req.PageSize, req.UserId)
	if err != nil {
		return nil, err
	}

	// 将订单视图列表转换为 RPC 响应格式
	resp := make([]*orderpb.ExchangeOrder, len(list))
	if err := copier.Copy(&resp, list); err != nil {
		return nil, err
	}
	return &orderpb.OrderRes{List: resp, Total: total}, nil
}
