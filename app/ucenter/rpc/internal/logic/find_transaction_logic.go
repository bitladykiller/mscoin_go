package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

type FindTransactionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFindTransactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTransactionLogic {
	return &FindTransactionLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *FindTransactionLogic) FindTransaction(req *assetpb.AssetReq) (*assetpb.MemberTransactionList, error) {
	list, total, err := l.svcCtx.TransactionService.FindTransaction(
		l.ctx,
		req.UserId,
		req.PageNo,
		req.PageSize,
		req.Symbol,
		req.StartTime,
		req.EndTime,
		req.Type,
	)
	if err != nil {
		return nil, err
	}

	return &assetpb.MemberTransactionList{
		List:  list,
		Total: total,
	}, nil
}
