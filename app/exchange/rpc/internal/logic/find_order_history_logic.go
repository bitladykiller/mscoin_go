package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// FindOrderHistoryLogic 处理查询历史订单的 RPC 请求。
// 返回用户已完成或已取消的历史订单列表。
type FindOrderHistoryLogic struct {
	exchangeLogicBase
}

// NewFindOrderHistoryLogic 创建 FindOrderHistoryLogic 实例。
func NewFindOrderHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindOrderHistoryLogic {
	return &FindOrderHistoryLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// FindOrderHistory 执行查询历史订单的业务逻辑。
//
// 处理流程：
// 1. 调用 OrderService.FindOrderHistory 查询数据库
// 2. 将订单实体列表转换为 RPC 响应格式
// 3. 返回分页结果
//
// 查询条件：
// - 用户 ID（memberId）
// - 交易对符号（symbol）
// - 分页参数（page, pageSize）
//
// 返回的订单状态：
// - COMPLETED: 已完成（订单全部成交）
// - CANCELED: 已取消（用户主动取消）
// - OVERTIMED: 已超时（订单超过有效期被系统取消）
//
// 注意：本方法不调用 ucenter-rpc 或 market-rpc，直接查询本地数据库。
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
