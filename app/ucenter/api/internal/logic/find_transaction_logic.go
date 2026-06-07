package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
	"mscoin_go/pkg/page"
)

type FindTransactionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindTransactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTransactionLogic {
	return &FindTransactionLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *FindTransactionLogic) FindTransaction(req *types.AssetReq) (*page.Result, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	userID := middleware.UserIDFromContext(l.ctx)
	payload, err := l.svcCtx.AssetClient.FindTransaction(ctx, &assetpb.AssetReq{
		UserId:    userID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		PageNo:    int64(pageNo),
		PageSize:  int64(pageSize),
		Symbol:    req.Symbol,
		Type:      req.Type,
	})
	if err != nil {
		return nil, err
	}

	items := make([]any, len(payload.List))
	for index, item := range payload.List {
		items[index] = item
	}

	return page.New(items, int64(pageNo), int64(pageSize), payload.Total), nil
}
