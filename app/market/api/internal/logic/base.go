package logic

import (
	"context"

	"mscoin_go/app/market/api/internal/svc"
)

type marketLogicBase struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func newMarketLogicBase(ctx context.Context, svcCtx *svc.ServiceContext) marketLogicBase {
	return marketLogicBase{ctx: ctx, svcCtx: svcCtx}
}
