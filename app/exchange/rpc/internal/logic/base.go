package logic

import (
	"context"

	"mscoin_go/app/exchange/rpc/internal/svc"
)

type exchangeLogicBase struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func newExchangeLogicBase(ctx context.Context, svcCtx *svc.ServiceContext) exchangeLogicBase {
	return exchangeLogicBase{ctx: ctx, svcCtx: svcCtx}
}
