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

type HistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HistoryLogic {
	return &HistoryLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *HistoryLogic) History(req *types.ExchangeReq) (*page.Result, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 10*time.Second)
	defer cancel()

	userID := middleware.UserIDFromContext(l.ctx)
	resp, err := l.svcCtx.OrderClient.FindOrderHistory(ctx, &orderpb.OrderReq{
		Symbol:   req.Symbol,
		Page:     req.PageNo,
		PageSize: req.PageSize,
		UserId:   userID,
	})
	if err != nil {
		return nil, err
	}

	items := make([]any, len(resp.List))
	for i := range resp.List {
		items[i] = resp.List[i]
	}
	return page.New(items, req.PageNo, req.PageSize, resp.Total), nil
}
