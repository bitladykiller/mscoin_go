// Package logic 定义了 exchange-api 的业务逻辑层。
// 每个 Logic 结构体负责处理特定的业务请求，调用 RPC 服务完成实际操作。
package logic

import (
	"context"
	"errors"

	"mscoin_go/app/exchange/api/internal/middleware"
	"mscoin_go/app/exchange/api/internal/svc"
	"mscoin_go/app/exchange/api/internal/types"
	orderpb "mscoin_go/app/exchange/rpc/pb/order"
)

// AddLogic 处理新增订单的业务逻辑。
type AddLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewAddLogic 创建 AddLogic 实例。
func NewAddLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddLogic {
	return &AddLogic{ctx: ctx, svcCtx: svcCtx}
}

// Add 执行新增订单的业务逻辑。
// 处理流程：
// 1. 验证请求参数的有效性
// 2. 从上下文获取用户 ID
// 3. 调用 RPC 服务创建订单
// 4. 返回订单 ID
func (l *AddLogic) Add(req *types.ExchangeReq) (string, error) {
	// 验证订单请求参数
	if !req.OrderValid() {
		return "", errors.New("invalid request")
	}

	// 从上下文获取已认证的用户 ID
	userID := middleware.UserIDFromContext(l.ctx)
	// 调用 RPC 服务创建订单
	resp, err := l.svcCtx.OrderClient.Add(l.ctx, &orderpb.OrderReq{
		Symbol:    req.Symbol,
		UserId:    userID,
		Direction: req.Direction,
		Type:      req.Type,
		Price:     req.Price,
		Amount:    req.Amount,
	})
	if err != nil {
		return "", err
	}
	return resp.OrderId, nil
}
