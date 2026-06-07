package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
	"mscoin_go/pkg/page"
)

type WithdrawRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawRecordLogic {
	return &WithdrawRecordLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *WithdrawRecordLogic) Record(req *types.WithdrawReq) (*page.Result, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	pageNo := req.Page
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	userID := middleware.UserIDFromContext(l.ctx)
	records, err := l.svcCtx.WithdrawClient.WithdrawRecord(ctx, &withdrawpb.WithdrawReq{
		UserId:   userID,
		Page:     int64(pageNo),
		PageSize: int64(pageSize),
	})
	if err != nil {
		return nil, err
	}

	items := make([]any, len(records.GetList()))
	for index, item := range records.GetList() {
		items[index] = item
	}

	return page.New(items, int64(pageNo), int64(pageSize), records.Total), nil
}
