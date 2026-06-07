package logic

import (
	"context"

	marketpb "mscoin_go/app/market/rpc/pb/market"
	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// WithdrawRecordLogic 提现记录查询逻辑处理器
type WithdrawRecordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewWithdrawRecordLogic 创建逻辑处理器实例
func NewWithdrawRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawRecordLogic {
	return &WithdrawRecordLogic{ctx: ctx, svcCtx: svcCtx}
}

// WithdrawRecord 查询会员提现记录
func (l *WithdrawRecordLogic) WithdrawRecord(req *withdrawpb.WithdrawReq) (*withdrawpb.RecordList, error) {
	list, total, err := l.svcCtx.WithdrawService.FindRecordList(l.ctx, req.UserId, req.Page, req.PageSize, func(ctx context.Context, coinID int64) (*marketpb.Coin, error) {
		return l.svcCtx.MarketClient.FindCoinById(ctx, &marketpb.MarketReq{Id: coinID})
	})
	if err != nil {
		return nil, err
	}

	return &withdrawpb.RecordList{
		List:  list,
		Total: total,
	}, nil
}
