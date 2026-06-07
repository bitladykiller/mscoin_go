package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
)

// exchangeLogicBase 是所有 Logic 结构体的基础类。
// 封装了公共的上下文和服务上下文引用，减少代码重复。
type exchangeLogicBase struct {
	// ctx 是请求上下文，用于传递超时和取消信号。
	ctx context.Context
	// svcCtx 是服务上下文，包含所有运行时依赖。
	svcCtx *svc.ServiceContext
}

// newExchangeLogicBase 创建 exchangeLogicBase 实例。
func newExchangeLogicBase(ctx context.Context, svcCtx *svc.ServiceContext) exchangeLogicBase {
	return exchangeLogicBase{ctx: ctx, svcCtx: svcCtx}
}
