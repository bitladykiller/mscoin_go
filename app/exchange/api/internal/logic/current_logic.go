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
// 处理流程：
// 1. 设置 10 秒超时上下文
// 2. 从上下文获取用户 ID
// 3. 调用 RPC 服务查询当前委托中的订单
// 4. 封装分页结果返回
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
