package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	assetpb "mscoin_go/app/ucenter/rpc/pb/asset"
)

// FindTransactionLogic 查询交易记录的逻辑处理器
type FindTransactionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewFindTransactionLogic 创建逻辑处理器实例
func NewFindTransactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindTransactionLogic {
	return &FindTransactionLogic{ctx: ctx, svcCtx: svcCtx}
}

// FindTransaction 查询会员交易记录
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
