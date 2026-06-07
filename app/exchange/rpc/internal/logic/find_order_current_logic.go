package logic

import (
	"context"

	"github.com/jinzhu/copier"

	"mscoin_go/app/exchange/rpc/internal/svc"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// FindOrderCurrentLogic 处理查询当前订单的 RPC 请求。
// 返回用户当前正在委托中的订单列表。
type FindOrderCurrentLogic struct {
	exchangeLogicBase
}

// NewFindOrderCurrentLogic 创建 FindOrderCurrentLogic 实例。
func NewFindOrderCurrentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindOrderCurrentLogic {
	return &FindOrderCurrentLogic{exchangeLogicBase: newExchangeLogicBase(ctx, svcCtx)}
}

// FindOrderCurrent 执行查询当前订单的业务逻辑。
//
// 处理流程：
// 1. 调用 OrderService.FindOrderCurrent 查询数据库
// 2. 将订单实体列表转换为 RPC 响应格式
// 3. 返回分页结果
//
// 查询条件：
// - 用户 ID（memberId）
// - 交易对符号（symbol）
// - 分页参数（page, pageSize）
// - 订单状态为 TRADING（交易中）
//
// 返回的订单状态：
// - TRADING: 交易中（订单正在撮合队列中等待成交）
//
// 注意：本方法不调用 ucenter-rpc 或 market-rpc，直接查询本地数据库。
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
