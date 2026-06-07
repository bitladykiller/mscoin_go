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

// CurrentLogic 处理查询当前订单的业务逻辑。
type CurrentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewCurrentLogic 创建 CurrentLogic 实例。
func NewCurrentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CurrentLogic {
	return &CurrentLogic{ctx: ctx, svcCtx: svcCtx}
}

// Current 执行查询当前订单的业务逻辑。
//
// 处理流程：
// 1. 设置 10 秒超时上下文（防止长时间阻塞）
// 2. 从上下文获取用户 ID（由认证中间件解析 JWT 后存入）
// 3. 调用 exchange-rpc 的 FindOrderCurrent 方法查询当前委托中的订单
// 4. 封装分页结果返回
//
// 与 exchange-rpc 的调用关系：
// - 本方法调用 OrderClient.FindOrderCurrent() 将查询请求转发到 exchange-rpc
// - exchange-rpc 会直接查询数据库，返回用户当前正在委托中的订单
// - 订单状态为 TRADING（交易中），表示订单正在撮合队列中等待成交
//
// 返回数据说明：
// - 返回分页结果，包含订单列表和总数
// - 每个订单包含：订单 ID、交易对、方向、类型、价格、数量、已成交数量、已成交金额等
// - 状态、方向、类型字段使用字符串标签（如 "TRADING"、"BUY"、"LIMIT_PRICE"），便于前端直接展示
func (l *CurrentLogic) Current(req *types.ExchangeReq) (*page.Result, error) {
	// 设置 10 秒超时，防止长时间阻塞
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	// 从上下文获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)
	// 调用 RPC 服务查询当前订单
	resp, err := l.svcCtx.OrderClient.FindOrderCurrent(ctx, &orderpb.OrderReq{
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
