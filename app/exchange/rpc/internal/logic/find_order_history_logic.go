package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// FindOrderHistoryLogic 处理查询历史订单的 RPC 请求。
type FindOrderHistoryLogic struct {
	exchangeLogicBase
}

// NewFindOrderHistoryLogic 创建 FindOrderHistoryLogic 实例。
func NewFindOrderHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindOrderHistoryLogic {
	return &FindOrderHistoryLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// FindOrderHistory 执行查询历史订单的业务逻辑。
// 返回用户的历史订单列表（已完成/已取消状态）。
func (l *FindOrderHistoryLogic) FindOrderHistory(req *orderpb.OrderReq) (*orderpb.OrderRes, error) {
	// 调用服务层查询历史订单
	list, total, err := l.svcCtx.OrderService.FindOrderHistory(l.ctx, req.Symbol, req.Page, req.PageSize, req.UserId)
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
