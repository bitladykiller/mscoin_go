package logic

import (
	"context"
	"time"

	"mscoin_go/app/ucenter/api/internal/middleware"
	"mscoin_go/app/ucenter/api/internal/svc"
	"mscoin_go/app/ucenter/api/internal/types"
	withdrawpb "mscoin_go/app/ucenter/rpc/pb/withdraw"
)

type WithdrawCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWithdrawCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WithdrawCodeLogic {
	return &WithdrawCodeLogic{ctx: ctx, svcCtx: svcCtx}
}

func (l *WithdrawCodeLogic) WithdrawCode(req *types.WithdrawReq) (string, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 5*time.Second)
	defer cancel()

	userID := middleware.UserIDFromContext(l.ctx)
	_, err := l.svcCtx.WithdrawClient.WithdrawCode(ctx, &withdrawpb.WithdrawReq{
		UserId:     userID,
		Unit:       req.Unit,
		JyPassword: req.JyPassword,
		Code:       req.Code,
		Address:    req.Address,
		Amount:     req.Amount,
		Fee:        req.Fee,
	})
	if err != nil {
		return "fail", err
	}
	return "success", nil
}
