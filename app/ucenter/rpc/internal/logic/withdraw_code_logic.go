package logic

import (
	"context"

	"mscoin_go/app/ucenter/rpc/internal/svc"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

// WithdrawCodeLogic 提现申请逻辑处理器
type WithdrawCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewWithdrawCodeLogic 创建逻辑处理器实例
func NewWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawCodeLogic {
	return &WithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

// WithdrawCode 处理提现申请
func (l *WithdrawCodeLogic) WithdrawCode(req *withdrawpb.WithdrawReq) (*withdrawpb.NoRes, error) {
	if err := l.svcCtx.WithdrawService.Apply(l.ctx, req); err != nil {
		return nil, err
	}
	return &withdrawpb.NoRes{}, nil
}
