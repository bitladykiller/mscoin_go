package logic

import (
	"context"
	"time"

	"mscoin_go/app/exchange/api/internal/middleware"
	"mscoin_go/app/exchange/api/internal/svc"
	"mscoin_go/app/exchange/api/internal/types"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
	"mscoin_go/pkg/page"
)

// HistoryLogic 处理查询历史订单的业务逻辑。
type HistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHistoryLogic 创建 HistoryLogic 实例。
func NewHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryLogic {
	return &HistoryLogic{ctx: ctx, svcCtx: svcCtx}
}

// History 执行查询历史订单的业务逻辑。
//
// 处理流程：
// 1. 设置 10 秒超时上下文（防止长时间阻塞）
// 2. 从上下文获取用户 ID（由认证中间件解析 JWT 后存入）
// 3. 调用 exchange-rpc 的 FindOrderHistory 方法查询历史订单
// 4. 封装分页结果返回
//
// 与 exchange-rpc 的调用关系：
// - 本方法调用 OrderClient.FindOrderHistory() 将查询请求转发到 exchange-rpc
// - exchange-rpc 会直接查询数据库，返回用户已完成或已取消的订单
//
// 返回数据说明：
// - 返回分页结果，包含订单列表和总数
// - 每个订单包含：订单 ID、交易对、方向、类型、价格、数量、已成交数量、已成交金额、完成时间/取消时间等
// - 状态、方向、类型字段使用字符串标签，便于前端直接展示
//
// 返回的订单状态：
// - COMPLETED: 已完成（订单全部成交）
// - CANCELED: 已取消（用户主动取消）
// - OVERTIMED: 已超时（订单超过有效期被系统取消）
func (l *HistoryLogic) History(req *types.ExchangeReq) (*page.Result, error) {
	// 设置 10 秒超时，防止长时间阻塞
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	// 从上下文获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)
	// 调用 RPC 服务查询历史订单
	resp, err := l.svcCtx.OrderClient.FindOrderHistory(ctx, &orderpb.OrderReq{
		Symbol:   req.Symbol,
		Page:     req.PageNo,
		PageSize: req.PageSize,
		UserId:   userID,
	})
	if err != nil {
		return nil, err
	}

	// 将 RPC 响应转换为通用分页结果
	items := make([]any, len(resp.List))
	for i := range resp.List {
		items[i] = resp.List[i]
	}
	return page.New(items, req.PageNo, req.PageSize, resp.Total), nil
}
